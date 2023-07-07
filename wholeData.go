package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type wholeData struct {
	apID       int
	domrfID    string
	project    string
	address    string
	manager    string
	rve        string
	stage      string
	stageDesc  string
	layer      int
	domrfDate1 string
	domrfDate2 string
}

var monthMap = map[string]string{"январь": "01", "февраль": "02", "март": "03", "апрель": "04", "май": "05", "июнь": "06", "июль": "07",
	"август": "08", "сентябрь": "09", "октябрь": "10", "ноябрь": "11", "декабрь": "12"}

func newData(aID, layer int, stage, stageDesc, drID, project, address, manager, rve, domrfDate1, domrfDate2 string) wholeData {
	return wholeData{apID: aID, domrfID: drID, project: project, address: address,
		manager: manager, rve: rve, stage: stage, stageDesc: stageDesc, layer: layer, domrfDate1: domrfDate1, domrfDate2: domrfDate2}
}

// Запись данных в эксель
func (wd wholeData) excelWriting(rowCounter int, book *excelize.File) error {
	var sh = "Sheet1"
	book.SetCellValue(sh, fmt.Sprintf("A%d", rowCounter), wd.apID)
	book.SetCellValue(sh, fmt.Sprintf("B%d", rowCounter), wd.domrfID)
	book.SetCellValue(sh, fmt.Sprintf("C%d", rowCounter), wd.project)
	book.SetCellValue(sh, fmt.Sprintf("D%d", rowCounter), wd.address)
	book.SetCellValue(sh, fmt.Sprintf("E%d", rowCounter), wd.manager)
	book.SetCellValue(sh, fmt.Sprintf("F%d", rowCounter), wd.rve)
	book.SetCellValue(sh, fmt.Sprintf("G%d", rowCounter), wd.stage)
	book.SetCellValue(sh, fmt.Sprintf("H%d", rowCounter), wd.stageDesc)
	book.SetCellValue(sh, fmt.Sprintf("I%d", rowCounter), wd.layer)
	book.SetCellValue(sh, fmt.Sprintf("J%d", rowCounter), wd.domrfDate1)
	book.SetCellValue(sh, fmt.Sprintf("K%d", rowCounter), wd.domrfDate2)
	a, err := wd.canClose()
	if err != nil {
		return err
	}
	book.SetCellValue(sh, fmt.Sprintf("L%d", rowCounter), a)

	return nil
}

// Логика расчета значения поля "Можно обновить"
func (wd wholeData) canClose() (bool, error) {
	if wd.layer == 0 {
		return false, nil
	}

	if wd.stage == "Введен в эксплуатацию" {
		return true, nil
	}

	currentDate := time.Now().Format("01.2006")
	rve, err := time.Parse("2006-01-02", wd.rve)
	if err != nil {
		return false, err
	}
	if currentDate == rve.Format("01.2006") {
		return false, nil
	}

	curDate := time.Now().Format("02.01.2006")
	tempCurrentDate, _ := time.Parse("02.01.2006", curDate)
	date2, err := time.Parse("02.01.2006", wd.domrfDate2)
	if err != nil {
		return false, err
	}
	if date2.AddDate(0, 0, 10).After(tempCurrentDate) || (date2.Month() == tempCurrentDate.Month() && date2.Year() == tempCurrentDate.Year()) {
		return true, nil
	}

	a := strings.Split(wd.domrfDate1, ", ")
	a[0] = monthMap[a[0]]
	if currentDate == strings.Join(a, ".") {
		return true, nil
	}

	return false, nil
}
