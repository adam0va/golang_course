package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/xml"
	"encoding/json"
	"os"
	"strings"
	"strconv"
	"time"
	"fmt"
	"sort"
	"io/ioutil"
)

const filePath string = "dataset.xml"

type DataItem struct {
	Id int `xml:"id" json:"id"`
	GUID string `xml:"guid" json:"guid"`
	IsActive bool `xml:"isActive" json:"isActive" `
	Balance string `xml:"balance" json:"balance"`
	Picture string `xml:"picture" json:"picture"`
	Age int `xml:"age" json:"age"`
	EyeColor string `xml:"eyeColor" json:"eyeColor"`
	FirstName string `xml:"first_name" xml:"first_name"`
	LastName string `xml:"last_name" json:"last_name"`
	Gender string `xml:"gender" json:"gender"`
	Company string `xml:"company" json:"company"`
	Email string `xml:"email" json:"email"`
	Phone string `xml:"phone" json:"phone"`
	Address string `xml:"address" json:"address"`
	About string `xml:"about" json:"about"`
	Registered string `xml:"registered" json:"registered"`
	FavoriteFruit string `xml:"favoriteFruit" json:"favoriteFruit"`
}

type DataItemsList struct {
	XMLName xml.Name   `xml:"root"`
	Items []*DataItem `xml:"row" json:"user"`
}

func GetDataFromXml() (items DataItemsList) {
	xmlFile, errOpen := os.Open(filePath) // открываем файл
	if errOpen != nil {
		panic(errOpen)
	}
	defer xmlFile.Close() // в конце - закрыть
	data, errorRead := ioutil.ReadAll(xmlFile) // читаем все из файла
	if errorRead != nil {
		panic(errorRead)
	}
	//fmt.Println(string(data))
	items.Items = make([]*DataItem, 0, 35)
	err := xml.Unmarshal(data, &items) // распаковаваем xml
	fmt.Println("HELLO 1")
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	//fmt.Println(items.Items[0].FirstName)
	return
}

func getRequestedDataFromXml(query string, data DataItemsList) ([]DataItem) {
	foundItems := make([]DataItem, 0)
	for _, item := range data.Items {
		if (strings.Contains(item.FirstName, query) || strings.Contains(item.LastName, query) || strings.Contains(item.About, query)) {
			foundItems = append(foundItems, *item)
		}
	}
	return foundItems
}

func handler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	itemsList := GetDataFromXml()
	// query - подстрока в Name или About для фильтрации структур
	query := r.URL.Query().Get("query")
	if query == "badJson" {
		w.Write([]byte("{"))
		return
	} else if query == "BadRequest" {
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(SearchErrorResponse{"Error"})
		w.Write(resp)
		return
	} else if query == "Error" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if query == "Empty" {
		return
	}
	items := getRequestedDataFromXml(query, itemsList)

	// orderfield - по какому параметру сортировать Name, Id, Age. Если пустой - по Name
	// если ничего из этого - ошибка
	order_field := r.URL.Query().Get("order_field")
	if order_field == "" {
		order_field = "Name"
	} else if (order_field != "Name" && order_field != "Id" && order_field != "Age") {
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	// orderby. -1, 0, 1
	order_by := r.URL.Query().Get("order_by")
	if order_by == "1" { // в обратном порядке
		if order_field == "Name" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].FirstName+items[i].LastName < items[j].FirstName+items[j].LastName
			})
		} else if order_field == "Id" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Id < items[j].Id
			})
		} else if order_field == "Age" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Age < items[j].Age
			})
		}
	} else if order_by == "-1" { // просто сортировать
		if order_field == "Name" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].FirstName+items[i].LastName > items[j].FirstName+items[j].LastName
			})
		} else if order_field == "Id" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Id > items[j].Id
			})
		} else if order_field == "Age" {
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Age > items[j].Age
			})
		}
	} else if order_by != "0" { // иначе ничего не сортируем
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit")) // limit
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("no limit:", err)
		return
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset")) // offset
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("no offset:", err)
		return
	}
	from := offset
	to := offset+limit
	if len(items) > limit {
		items = items[from:to]
	}
	//fmt.Println(items)
	responseData, err := json.Marshal(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("cant pack result json:", err)
		return
	}
	w.Write(responseData)
}

// ---- TEST ----

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	cases := []SearchRequest{
		SearchRequest{
			Limit: 10,
			Offset: 10,
			Query: "",
			OrderField: "Name",
			OrderBy: 1,
		},
		SearchRequest{
			Limit: 1,
			Offset: 10,
			Query: "",
			OrderField: "Name",
			OrderBy: 1,
		},
		SearchRequest{
			Limit: 25,
			Offset: 1,
			OrderField: "Name",
			OrderBy: 1,
		},
		SearchRequest{
			Limit: 23,
			Query: "Guerra",
			OrderField: "Name",
			OrderBy: 1,
		},
	}
	for _, testCase := range cases {
		_, err := testClient.FindUsers(testCase)
		if err != nil {
			t.Errorf("Some error")
		}
	}
}

func TestUnauthrized403(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	testClient := SearchClient{
		AccessToken : "",
		URL : ts.URL,
	}

	cases := []SearchRequest{
		SearchRequest{
			Limit: 10,
			Offset: 10,
			Query: "",
			OrderField: "Name",
			OrderBy: 1,
		},
	}
	for _, testCase := range cases {
		resp, err := testClient.FindUsers(testCase)
		if resp != nil || err.Error() != "Bad AccessToken" {
			t.Errorf("403 doesn't work")
		}
	}
}

func TestInternal500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))


	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	testcase500 := SearchRequest{
					Query: "",
					OrderField: "Name",
					OrderBy: 5,
				}
		
	_, err := testClient.FindUsers(testcase500)
	if err.Error() != "SearchServer fatal error" {
		t.Errorf("500 doesn't work")
	}
}

func TestUnknown(t *testing.T) {
	testClient := SearchClient{
		AccessToken : "0",
		URL : "",
	}

	resp, err := testClient.FindUsers(SearchRequest{})
	if resp != nil && err == nil {
		t.Errorf("Unknown error")
	}	
}

func TestBadLimitAndOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	case1 := 	SearchRequest{
					Limit: -1,
					Offset: 10,
					Query: "",
					OrderField: "Name",
					OrderBy: 1,
				}

	resp, err := testClient.FindUsers(case1)
	if resp != nil && err.Error() != "limit must be > 0" {
		t.Errorf("bad limit doesn't work")
	}

	case2 := 	SearchRequest{
					Limit: 10,
					Offset: -1,
					Query: "",
					OrderField: "Name",
					OrderBy: 1,
				}

	resp, err = testClient.FindUsers(case2)
	if resp != nil && err.Error() != "offset must be > 0" {
		t.Errorf("bad offset doesn't work")
	}

	case3 := 	SearchRequest{
					Limit: 26,
					Offset: 10,
					Query: "",
					OrderField: "Name",
					OrderBy: 1,
				}

	resp, err = testClient.FindUsers(case3)
	if err != nil || resp == nil {
		t.Errorf("limit change doesn't work")
	}
	if (len(resp.Users) > 25) {
		t.Errorf("limit change doesn't work (limit > 25)")
	}
}

func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5*time.Second)
}

func TestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(timeoutHandler))

	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	timeoutCase := 	SearchRequest{
						Limit: 10,
						Offset: 10,
						Query: "",
						OrderField: "Name",
						OrderBy: 1,
					}
	_, err := testClient.FindUsers(timeoutCase)
	if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("timeout doesn't work")
	}
	
}

func TestBadOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	timeoutCase := 	SearchRequest{
						Limit: 10,
						Offset: 10,
						Query: "",
						OrderField: "Field",
						OrderBy: 1,
					}
	_, err := testClient.FindUsers(timeoutCase)
	if !strings.Contains(err.Error(), "OrderFeld") {
		t.Errorf("Bad order field doesn't work")
	}
	
}

func TestBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	testClient := SearchClient{
		AccessToken : "0",
		URL : ts.URL,
	}

	testCases := 	[]SearchRequest {
						SearchRequest{
							Limit: 10,
							Offset: 10,
							Query: "BadRequest",
							OrderField: "Name",
							OrderBy: 1,
						},
						SearchRequest{
							Limit: 10,
							Offset: 10,
							Query: "Error",
							OrderField: "Name",
							OrderBy: 1,
						},
						SearchRequest {
							Limit: 10,
							Offset: 10,
							Query: "Empty",
							OrderField: "Name",
							OrderBy: 1,
						},
					}
	// if !strings.Contains(err.Error(), "cant unpack error json:") {
	// 	t.Errorf("General bad request doesn't work")
	// }
	for _, testCase := range testCases {
		_, err := testClient.FindUsers(testCase)
		fmt.Printf("badddd %v\n", err.Error())
		if err == nil {
			t.Errorf("Test doen't work")
		}
	}
	
}























