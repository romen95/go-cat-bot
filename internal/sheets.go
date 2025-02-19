package internal

import (
	"context"
	"fmt"

	"gocarbot/models"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func IsOrderEmpty(order models.Order) bool {
	return order.OrderNumber == "" &&
		order.OrderDate == "" &&
		order.EstimatedArrival == "" &&
		order.CarModel == "" &&
		order.CountryOfOrigin == "" &&
		order.CurrentLocation == "" &&
		order.Destination == ""
}

func GetUserByID(userID string) (*models.User, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("internal/credentials.json"))
	if err != nil {
		return nil, err
	}

	const spreadsheetID = "1Y3Lj-7XhiP03ojyhmjVEukqK8mHr7bohuZMrMwUSYN8"
	const rangeName = "Users!A:Q"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range resp.Values {
		if len(row) < 1 || fmt.Sprintf("%v", row[0]) != userID {
			continue
		}

		user := &models.User{
			ID:          safeString(row, 0),
			PhoneNumber: safeString(row, 1),
			Order1: models.Order{
				OrderNumber:      safeString(row, 2),
				OrderDate:        safeString(row, 3),
				EstimatedArrival: safeString(row, 4),
				CarModel:         safeString(row, 5),
				CountryOfOrigin:  safeString(row, 6),
				CurrentLocation:  safeString(row, 7),
				Destination:      safeString(row, 8),
			},
			Order2: models.Order{
				OrderNumber:      safeString(row, 9),
				OrderDate:        safeString(row, 10),
				EstimatedArrival: safeString(row, 11),
				CarModel:         safeString(row, 12),
				CountryOfOrigin:  safeString(row, 13),
				CurrentLocation:  safeString(row, 14),
				Destination:      safeString(row, 15),
			},
		}
		return user, nil
	}

	return nil, fmt.Errorf("пользователь с ID %s не найден", userID)
}

// safeString - безопасное преобразование значения к строке, если оно существует.
func safeString(row []interface{}, index int) string {
	if index < len(row) {
		if str, ok := row[index].(string); ok {
			return str
		}
		return fmt.Sprintf("%v", row[index])
	}
	return ""
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
	const rangeName = "Users!A:B" // Столбцы: A - ID, B - Телефон

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
