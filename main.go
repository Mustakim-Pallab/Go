package main

import (
	"fmt"
	"net/http"
	"sort"
	// "strconv"

	// "errors"
	_ "github.com/go-sql-driver/mysql"
	echo "github.com/labstack/echo/v4"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// DB connection
	var dbErr error
	dsn := "root:12345@tcp(localhost:3306)/pallab1?charset=utf8mb4&parseTime=True&loc=Local"
	db, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if dbErr != nil {
		panic("falied to connect to db!")
	}

	// new instance of the app
	e := echo.New()
	fmt.Println(db)

	// routes
	e.GET("/", hello)
	e.GET("/activities", func(c echo.Context) error {
		return getActivities(c, db)
	})
	

	
	// starting server
	err := e.Start(":8080")
	if err != nil {
		panic(err)
	}
}

// returns hello world if the corresponding request is sent
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

type User struct {
	Id             uint64 `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
}

type Activitiy struct {
	Id             uint   `json:"id"`
	FirstName      string `json:"first_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
	Points         uint   `json: "points"`
	Rank           int    `json:"rank"`
}


// returns an array of users containing points according to their activity
func getActivities(c echo.Context, db *gorm.DB) error {
	var useractivities []Activitiy
	resp := db.Raw("SELECT users.id, users.first_name, users.country, users.profile_picture, activities.points FROM ((users INNER JOIN activity_logs ON users.id = activity_logs.user_id) INNER JOIN activities ON activity_logs.activity_id = activities.id)").Find(&useractivities)
	if resp.Error != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	// calculatePoints() is used to calculate all the points of each user
	
	pointsMap := calculatePoints(useractivities)
	for i := range useractivities {
		useractivities[i].Points = pointsMap[useractivities[i].Id] // total points of a user is being assigned to the main array
	}

	// removeDuplicates() removes the duplicate elements of the responed array of db and maps each user's achieved point acccording to id
	uniqueUsers := removeDuplicates(useractivities)

	// Sorted useractivities based on points in descending order
	sort.SliceStable(uniqueUsers, func(i, j int) bool {
		return uniqueUsers[i].Points > uniqueUsers[j].Points
	})

	// raneked the sorted array
	rank := 1
	uniqueUsers[0].Rank = rank
	for i := 1; i < len(uniqueUsers); i++ {
		if uniqueUsers[i].Points < uniqueUsers[i-1].Points {
			rank++ // rank of each users are decided by comparing with previous user
		}
		uniqueUsers[i].Rank = rank
	}

	return c.JSON(http.StatusOK, uniqueUsers)
}

// calculatePoints() is used to calculate all the points of each user

func calculatePoints(users []Activitiy) map[uint]uint {
	pointsMap := make(map[uint]uint)
	for _, user := range users {
		pointsMap[user.Id] += user.Points // points are calculated for single user
	}
	return pointsMap
}

// removeDuplicates() removes the duplicate elements of the responed array of db and maps each user's achieved point acccording to id
func removeDuplicates(users []Activitiy) []Activitiy {
	uniqueUsers := make(map[uint]Activitiy)
	for _, user := range users {
		uniqueUsers[user.Id] = user // overwrites multiple instances with the updated one
	}

	result := make([]Activitiy, 0, len(uniqueUsers))
	for _, user := range uniqueUsers {
		result = append(result, user) // appends the user to the new slice result
	}
	return result
}