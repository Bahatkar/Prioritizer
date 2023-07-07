package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"prioritizer/repository"

	"github.com/gocolly/colly"
	"github.com/spf13/viper"
)

type building struct {
	Properties struct {
		InitState struct {
			KN struct {
				ObjCard struct {
					ConstrProg struct {
						Info []struct {
							Src       string `json:"src"`
							LocalDate string `json:"localDate"`
							Date      string `json:"dateOfPlacement"`
						} `json:"shortInfo"`
					} `json:"constructionProgress"`
				} `json:"objectCard"`
			} `json:"kn"`
		} `json:"initialState"`
	} `json:"props"`
}

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("Error occured while initializing configs: %s", err)
	}

	dbNameCollection := viper.GetStringSlice("ap.dbname")
	//запуск цикла по каждой базе из слайса
	for _, dbName := range dbNameCollection {
		db, err := repository.NewMySQLDB(repository.Config{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetString("db.port"),
			Username: viper.GetString("db.username"),
			Protocol: viper.GetString("db.protocol"),
			DBName:   dbName,
			Password: viper.GetString("db.password"),
		})
		if err != nil {
			log.Fatalf("Error occured while opening DB: %s", err)
		}
		defer db.Close()

		var cycleID int
		rows, err := db.Query(`select max(pc.id) from parser_cycle pc`)
		if err != nil {
			log.Println(err)
		}
		rows.Scan()
		for rows.Next() {
			rows.Scan(&cycleID)
		}

		//Запрос на получение данных из бд
		rows, err = db.Query(`
		SELECT
		*
		FROM(
			SELECT 
				pcb.building_id, 
				pb.bnmap_id, 
				phc.name AS project, 
				pb.name, 
				CONCAT(au.firstname,' ',au.lastname),
				pbsh.date_state_commission,
				ps.stage,
				pbsh.stage_desc,
				(CASE WHEN (SELECT pbc.id FROM parser_building_collection pbc WHERE pbc.building_id = pcb.building_id 
				AND YEAR(pbc.createTimeMax) = YEAR(CURDATE()) AND MONTH(pbc.createTimeMax) = MONTH(CURDATE()) ORDER BY pbc.createTimeMax DESC LIMIT 1)
				IS NULL THEN 0 ELSE 1 END),
				ROW_NUMBER() OVER (PARTITION BY pbsh.building_id ORDER BY pbsh.create_time DESC) AS nmb			
			FROM 
				parser_cycle_building pcb
				LEFT JOIN parser_building pb ON pcb.building_id = pb.id
				LEFT JOIN parser_housing_complex phc ON pb.housing_complex_id = phc.id
				LEFT JOIN auth_user au ON pcb.user_id = au.id
				LEFT JOIN parser_building_stage_history pbsh ON pbsh.building_id = pcb.building_id
				LEFT JOIN parser_stage ps ON ps.id = pbsh.stage_id
			WHERE 
				pcb.cycle_id = ? 
				AND pcb.status = 0
				AND pb.sale_status = 1) tmp
		WHERE tmp.nmb = 1`, cycleID)
		if err != nil {
			log.Println(err)
		}

		excelFile := newBook()
		var rowCounter = 1

		for rows.Next() {
			rowCounter++
			var (
				buildingId, gotLayer, rowNumber                            int
				project, address, userName, bnmapId, rve, stage, stageDesc sql.NullString
			)
			err = rows.Scan(&buildingId, &bnmapId, &project, &address, &userName, &rve, &stage, &stageDesc, &gotLayer, &rowNumber)
			if err != nil {
				break
			}
			//Проверка полученных из бд данных на null значения и их замена
			if !project.Valid {
				project.String = "n/a"
			}
			if !address.Valid {
				address.String = "n/a"
			}
			if !userName.Valid {
				userName.String = "n/a"
			}
			if !bnmapId.Valid {
				bnmapId.String = "n/a"
			}
			if !stage.Valid {
				stage.String = "n/a"
			}
			if !stageDesc.Valid {
				stage.String = "n/a"
			}
			if !rve.Valid {
				rve.String = "1992-06-04"
			}
			log.Printf("Собран корпус %s-%d\n", dbName, buildingId)

			//Получение данных с дом.рф
			houseNumber := strings.Split(bnmapId.String, ",")[0]
			date1, date2, err := getDate(houseNumber)
			if err != nil {
				log.Printf("Error occured while parsing dom.rf ID %s: %s\n", bnmapId.String, err)
			}

			data := newData(buildingId, gotLayer, stage.String, stageDesc.String, bnmapId.String, project.String, address.String, userName.String,
				rve.String, date1, date2)

			err = data.excelWriting(rowCounter, excelFile.book)
			if err != nil {
				log.Fatalf("Error occured while writing Excel: %s\n", err)
			}
		}

		excelFile.close(dbName)
	}
}

// Получение конфигов из .yaml файла
func initConfig() error {
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

// Парсинг данных с дом.рф
func getDate(houseNumber string) (date, date1 string, err error) {
	url := "https://xn--80az8a.xn--d1aqf.xn--p1ai/%D1%81%D0%B5%D1%80%D0%B2%D0%B8%D1%81%D1%8B/%D0%BA%D0%B0%D1%82%D0%B0%D0%BB%D0%BE%D0%B3-" +
		"%D0%BD%D0%BE%D0%B2%D0%BE%D1%81%D1%82%D1%80%D0%BE%D0%B5%D0%BA/%D0%BE%D0%B1%D1%8A%D0%B5%D0%BA%D1%82/"
	c := colly.NewCollector()

	var b = &building{}
	c.OnHTML("#__NEXT_DATA__", func(h *colly.HTMLElement) {
		json.Unmarshal([]byte(h.Text), b)
	})

	c.Visit(url + houseNumber)

	shortInfo := b.Properties.InitState.KN.ObjCard.ConstrProg.Info
	date = "июнь, 1992"
	date1 = "04.06.1992"
	if len(shortInfo) != 0 {
		date = shortInfo[len(shortInfo)-1].LocalDate
		date1 = shortInfo[len(shortInfo)-1].Date
	}

	return date, date1, nil
}
