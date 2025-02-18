package internal

import (
	"context"
	"fmt"

	"gocarbot/models"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// getUserByID - функция для получения данных пользователя по ID из Google Sheets
func GetUserByID(userID string) (*models.User, error) {
	// Создаем клиент для работы с Google Sheets API
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("internal/credentials.json"))
	if err != nil {
		return nil, err
	}

	// ID таблицы и диапазон данных
	const spreadsheetID = "1Y3Lj-7XhiP03ojyhmjVEukqK8mHr7bohuZMrMwUSYN8"
	const rangeName = "Users!A:C" // Столбцы: A - ID, B - Номер телефона, C - Автомобиль

	// Получаем данные из таблицы
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return nil, err
	}

	// Перебираем все строки и ищем нужного пользователя
	for _, row := range resp.Values {
		if len(row) >= 3 && row[0] == userID {
			// Заполняем структуру User
			user := &models.User{
				ID:          row[0].(string),
				PhoneNumber: row[1].(string),
				Car:         row[2].(string),
			}
			return user, nil
		}
	}

	// Если пользователь не найден
	return nil, fmt.Errorf("пользователь с ID %s не найден", userID)
}

func RequestPhoneAndUpdateID(userID, phoneNumber string) error {
	// Создаем клиент для работы с Google Sheets API
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("internal/credentials.json"))
	if err != nil {
		return fmt.Errorf("ошибка при инициализации Google Sheets API: %v", err)
	}

	// ID таблицы и диапазон данных
	const spreadsheetID = "1Y3Lj-7XhiP03ojyhmjVEukqK8mHr7bohuZMrMwUSYN8"
	const rangeName = "Users!A:B" // Столбцы: A - ID, B - Номер телефона

	// Получаем данные из таблицы
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return fmt.Errorf("ошибка при получении данных: %v", err)
	}

	var rowIndex int
	var userRow []interface{}
	found := false

	// Перебираем строки и ищем номер телефона
	for index, row := range resp.Values {
		if len(row) >= 2 && row[1] == phoneNumber {
			// Если номер телефона найден
			rowIndex = index + 1 // Строки в Sheets начинаются с 1

			// Проверяем, есть ли уже ID в ячейке
			if len(row) > 0 && row[0] != "" {
				return fmt.Errorf("для данного номера уже задан ID: %v", row[0])
			}

			// Формируем новую строку для обновления
			userRow = []interface{}{userID, row[1]} // ID пользователя и телефон
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("номер не найден в базе")
	}

	// Обновляем данные в Google Sheets
	updateRange := fmt.Sprintf("Users!A%d:B%d", rowIndex, rowIndex)
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, updateRange, &sheets.ValueRange{
		Values: [][]interface{}{userRow},
	}).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return fmt.Errorf("ошибка при обновлении данных: %v", err)
	}

	return nil
}
