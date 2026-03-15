package application

import (
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	socks5Version      = 0x05
	authMethodNone     = 0x00
	authMethodUserPass = 0x02
	cmdConnect         = 0x01
	addrTypeIPv4       = 0x01
	addrTypeDomain     = 0x03
	addrTypeIPv6       = 0x04
)

// authenticate Аутентификация: проверка логина и пароля
func authenticate(conn net.Conn, app *App) error {
	// 1. Читаем заголовок приветствия (Версия + кол-во методов)
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}

	version := buf[0]
	if version != socks5Version {
		return errors.New("unsupported SOCKS version")
	}

	nMethods := int(buf[1])
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}

	// 2. Логика выбора метода
	if !app.Config.UseAuth {
		// Если аутентификация выключена, выбираем NO AUTH (0x00)
		// Важно: мы просто отправляем ответ и возвращаем nil.
		// Мы НЕ читаем логин/пароль и НЕ закрываем соединение.
		_, err := conn.Write([]byte{socks5Version, authMethodNone})
		return err // Возвращаем ошибку записи, если она есть, иначе nil
	}

	// Если аутентификация включена, проверяем, поддерживает ли клиент USER/PASS
	supportsUserPass := false
	for _, m := range methods {
		if m == authMethodUserPass {
			supportsUserPass = true
			break
		}
	}

	if !supportsUserPass {
		// Если сервер требует auth, а клиент не предложил нужный метод -> отключаем (0xFF)
		_, _ = conn.Write([]byte{socks5Version, 0xFF})
		return errors.New("no supported authentication method")
	}

	// Сообщаем клиенту, что выбрали метод USER/PASS
	if _, err := conn.Write([]byte{socks5Version, authMethodUserPass}); err != nil {
		return err
	}

	// 3. Читаем данные аутентификации (только если мы здесь, значит auth включен)
	authHeader := make([]byte, 2)
	if _, err := io.ReadFull(conn, authHeader); err != nil {
		return err
	}

	if authHeader[0] != 0x01 {
		return errors.New("invalid auth version")
	}

	usernameLen := int(authHeader[1])
	username := make([]byte, usernameLen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return err
	}

	passLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, passLenBuf); err != nil {
		return err
	}

	passwordLen := int(passLenBuf[0])
	password := make([]byte, passwordLen)
	if _, err := io.ReadFull(conn, password); err != nil {
		return err
	}

	// 4. Проверка учетных данных
	user, _ := app.UserStorage.FindByUsername(string(username))
	if user == nil || user.Password != string(password) {
		// Неверный логин/пароль
		// 0x01 0x01 означает: версия 0x01, статус FAILURE
		_, _ = conn.Write([]byte{0x01, 0x01})
		return errors.New("invalid credentials")
	}

	// Успех
	// 0x01 0x00 означает: версия 0x01, статус SUCCESS
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return err
	}

	return nil
}

// HandleRequest Обработка запроса клиента
func HandleRequest(conn net.Conn, app *App) error {
	defer conn.Close()

	// Аутентификация
	err := authenticate(conn, app)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Читаем запрос клиента
	request := make([]byte, 4)
	_, err = io.ReadFull(conn, request)
	if err != nil {
		return err
	}

	version := request[0]
	cmd := request[1]
	addrType := request[3]

	if version != socks5Version || cmd != cmdConnect {
		return errors.New("unsupported request")
	}

	var targetAddr string

	switch addrType {
	case addrTypeIPv4:
		ip := make([]byte, 4)
		_, err = io.ReadFull(conn, ip)
		if err != nil {
			return err
		}
		targetAddr = net.IP(ip).String()

	case addrTypeDomain:
		domainLenBuf := make([]byte, 1)
		_, err = io.ReadFull(conn, domainLenBuf)
		if err != nil {
			return err
		}
		domainLen := int(domainLenBuf[0])
		domain := make([]byte, domainLen)
		_, err = io.ReadFull(conn, domain)
		if err != nil {
			return err
		}
		targetAddr = string(domain)

	case addrTypeIPv6:
		ip := make([]byte, 16)
		_, err = io.ReadFull(conn, ip)
		if err != nil {
			return err
		}
		targetAddr = net.IP(ip).String()

	default:
		return errors.New("unsupported address type")
	}

	// Читаем порт (2 байта, big-endian)
	portBuf := make([]byte, 2)
	_, err = io.ReadFull(conn, portBuf)
	if err != nil {
		return err
	}
	port := int(portBuf[0])<<8 | int(portBuf[1])
	targetAddr = fmt.Sprintf("%s:%d", targetAddr, port)

	// Подключаемся к целевому серверу
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		// Отправляем ошибку (0x04 = Host unreachable)
		conn.Write([]byte{socks5Version, 0x04, 0x00, addrTypeIPv4, 0, 0, 0, 0, 0, 0})
		return fmt.Errorf("failed to connect to target: %v", err)
	}
	defer targetConn.Close()

	// Отправляем успешный ответ (0x00 = success)
	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	response := []byte{socks5Version, 0x00, 0x00}
	if localAddr.IP.To4() != nil {
		response = append(response, addrTypeIPv4)
		response = append(response, localAddr.IP.To4()...)
	} else {
		response = append(response, addrTypeIPv6)
		response = append(response, localAddr.IP.To16()...)
	}
	response = append(response, byte(localAddr.Port>>8), byte(localAddr.Port&0xFF))
	_, err = conn.Write(response)
	if err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	// Перенаправляем трафик между клиентом и целевым сервером
	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)

	return nil
}
