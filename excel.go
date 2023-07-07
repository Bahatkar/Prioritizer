package main

import (
	"fmt"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type excel struct {
	book *excelize.File
}

func newBook() excel {
	var rusCols = []string{"Айди АП", "Айди дом.рф", "Проект", "Адрес", "Специалист", "Дата РВЭ", "Стадия", "Описание стадии", "Наличие слоя",
		"Месяц обновления", "Дата фото", "Можно обновить"}

	f := excelize.NewFile()
	for i := 0; i < len(rusCols); i++ {
		cellRus := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cellRus, rusCols[i])
	}

	return excel{book: f}
}

func (e excel) close(dbName string) error {
	date := time.Now().Format("02.01.2006")
	fileName := fmt.Sprintf("%s_%s.%s", dbName, date, "xlsx")
	err := e.book.SaveAs(fileName)

	return err
}
