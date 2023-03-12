package storewithteacher

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Store struct {
	Id       int64
	Name     string
	Branches []*Branch
}

type Branch struct {
	Id           int64
	Name         string
	PhoneNumbers []string
	Address      *Address
	Vacancies    []*Vacancy
}

type Vacancy struct {
	Id     int64
	Name   string
	Salary float64
}

type Address struct {
	Id         int64
	City       string
	StreetName string
}

type Response struct {
	Store []*Store
}

func StoreTeacher() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "bnnfav", "store_gorm")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := []Store{
		{
			Name: "Nike",
			Branches: []*Branch{
				{
					Name: "Jordan",
					PhoneNumbers: []string{
						"998977777777",
						"998988888888",
					},
					Address: &Address{
						City:       "New York",
						StreetName: "Time Square",
					},
					Vacancies: []*Vacancy{
						{
							Name:   "Cleaner",
							Salary: 2300,
						},
						{
							Name:   "Marketolog",
							Salary: 9800,
						},
					},
				},
				{
					Name: "Air",
					PhoneNumbers: []string{
						"998999999999",
						"998900000000",
					},
					Address: &Address{
						City:       "Las-Vegas",
						StreetName: "Muccino",
					},
					Vacancies: []*Vacancy{
						{
							Name:   "Saler",
							Salary: 3500,
						},
						{
							Name:   "Manager",
							Salary: 6000,
						},
					},
				},
			},
		},
		{
			Name: "Adidas",
			Branches: []*Branch{
				{
					Name: "Merci",
					PhoneNumbers: []string{
						"22355555555",
						"22344444444",
					},
					Address: &Address{
						City:       "London",
						StreetName: "Big Ben",
					},
					Vacancies: []*Vacancy{
						{
							Name:   "Cleaner",
							Salary: 1800,
						},
						{
							Name:   "Designer",
							Salary: 4000,
						},
					},
				},
				{
					Name: "Sevou",
					PhoneNumbers: []string{
						"22366666666",
						"22365655556",
					},
					Address: &Address{
						City:       "Amsterdam",
						StreetName: "Oslo",
					},
					Vacancies: []*Vacancy{
						{
							Name:   "Casher",
							Salary: 2200,
						},
						{
							Name:   "Manager",
							Salary: 4900,
						},
					},
				},
			},
		},
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("error beginning transaction: ", err)
		return
	}

	for _, store := range store {
		var storeId int64
		err := tx.QueryRow("insert into stores(name) values($1) returning id", store.Name).Scan(&storeId)
		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			return
		}

		for _, branch := range store.Branches {
			var branchId int64
			err := tx.QueryRow("insert into branches(name, phone_numbers, store_id) values($1, $2, $3) returning id", branch.Name, pq.Array(branch.PhoneNumbers), storeId).Scan(&branchId)
			if err != nil {
				fmt.Println(err)
				tx.Rollback()
				return
			}

			_, err = tx.Exec("insert into address (city, street_name, branch_id) values($1, $2, $3)", branch.Address.City, branch.Address.StreetName, branchId)
			if err != nil {
				fmt.Println(err)
				tx.Rollback()
				return
			}

			for _, vacancy := range branch.Vacancies {
				var vacancyId int64

				err = tx.QueryRow("insert into vacancies (name, salary) values($1, $2) returning id", vacancy.Name, vacancy.Salary).Scan(&vacancyId)
				if err != nil {
					fmt.Println(err)
					tx.Rollback()
					return
				}

				_, err = tx.Exec("insert into branches_vacancies(branch_id, vacancy_id) values($1, $2)", branchId, vacancyId)
				if err != nil {
					fmt.Println(err)
					tx.Rollback()
					return
				}

			}
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("error commiting transaction: ", err)
		tx.Rollback()
		return
	}

	
	resp := Response{}

	storeRow, err := db.Query("select id, name from stores")
	if err != nil {
		fmt.Println(err)
		return
	}
	for storeRow.Next() {
		store := Store{}
		err = storeRow.Scan(
			&store.Id,
			&store.Name,
		)
		if err != nil {
			fmt.Println(err)
			return
		}

		branchRow, err := db.Query("select id, name, phone_number from branches where store_id = $1", store.Id)
		if err != nil {
			fmt.Println(err)
			return
		}

		for branchRow.Next() {
			branch := Branch{}
			err = branchRow.Scan(
				&branch.Id,
				&branch.Name,
				pq.Array(branch.PhoneNumbers),
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			vacancyRow, err := db.Query(`select v.id, v.name, v.salary from vacancies v join branches_vacancies br on v.id = br.vacancy_id join branches b on b.id = br.branch_id where b.id = $1`, branch.Id)
			if err != nil {
				fmt.Println(err)
				return
			}

			for vacancyRow.Next() {
				vacancy := Vacancy{}
				err = vacancyRow.Scan(
					&vacancy.Id,
					&vacancy.Name,
					&vacancy.Salary,
				)
				if err != nil {
					fmt.Println(err)
					return
				}

				branch.Vacancies = append(branch.Vacancies, &vacancy)

				addressRow, err := db.Query("select id, city, street_name from address where branch_id = $1", branch.Id)
				if err != nil {
					fmt.Println(err)
					return
				}

				for addressRow.Next() {
					address := Address{}
					err = addressRow.Scan(
						&address.Id,
						&address.City,
						&address.StreetName,
					)
					if err != nil {
						fmt.Println(err)
						return
					}

					branch.Address = &address
				}

			}

			store.Branches = append(store.Branches, &branch)
		}

		resp.Store = append(resp.Store, &store)
	}

	for _, store := range resp.Store {
		fmt.Println(store)

		for _, branch := range store.Branches {
			fmt.Println(branch)

			for _, vacancy := range branch.Vacancies {
				fmt.Println(vacancy)
				fmt.Println(branch.Address)
			}
		}
	}
	

}
