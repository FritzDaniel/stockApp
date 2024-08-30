package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var log = logrus.New()

// Stock struct
type Stock struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaBarang     string         `json:"nama_barang"`
	JumlahStok     int            `json:"jumlah_stok"`
	NomorSeri      string         `json:"nomor_seri"`
	AdditionalInfo datatypes.JSON `json:"additional_info"`
	GambarBarang   string         `json:"gambar_barang"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      string         `json:"created_by"`
	UpdatedAt      time.Time      `json:"updated_at"`
	UpdatedBy      string         `json:"updated_by"`
}

func InitDB() {
	var err error
	dsn := "host=localhost user=postgres password=postgres dbname=stockApp port=5432 sslmode=disable"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	DB.AutoMigrate(&Stock{})
}

func init() {
	// Logging setup
	log.Out = os.Stdout
	file, err := os.OpenFile("requests.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func main() {
	InitDB()

	r := mux.NewRouter()
	r.HandleFunc("/stock", CreateStock).Methods("POST")
	r.HandleFunc("/stock", ListStock).Methods("GET")
	r.HandleFunc("/stock/{id}", GetStockDetail).Methods("GET")
	r.HandleFunc("/stock/{id}", UpdateStock).Methods("PUT")
	r.HandleFunc("/stock/{id}", DeleteStock).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock Stock
	json.NewDecoder(r.Body).Decode(&stock)
	stock.CreatedAt = time.Now()

	result := DB.Create(&stock)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.WithFields(logrus.Fields{"action": "create", "data": stock}).Info("Create stock API called")
	json.NewEncoder(w).Encode(stock)
}

func ListStock(w http.ResponseWriter, r *http.Request) {
	var stocks []Stock
	DB.Find(&stocks)
	log.WithFields(logrus.Fields{"action": "list", "count": len(stocks)}).Info("List stock API called")
	json.NewEncoder(w).Encode(stocks)
}

func GetStockDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stock Stock
	if err := DB.First(&stock, vars["id"]).Error; err != nil {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}
	log.WithFields(logrus.Fields{"action": "detail", "id": vars["id"]}).Info("Detail stock API called")
	json.NewEncoder(w).Encode(stock)
}

func UpdateStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stock Stock
	if err := DB.First(&stock, vars["id"]).Error; err != nil {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}
	json.NewDecoder(r.Body).Decode(&stock)
	stock.UpdatedAt = time.Now()

	DB.Save(&stock)
	log.WithFields(logrus.Fields{"action": "update", "id": vars["id"]}).Info("Update stock API called")
	json.NewEncoder(w).Encode(stock)
}

func DeleteStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var stock Stock
	if err := DB.First(&stock, vars["id"]).Error; err != nil {
		http.Error(w, "Stock not found", http.StatusNotFound)
		return
	}
	DB.Delete(&stock)
	log.WithFields(logrus.Fields{"action": "delete", "id": vars["id"]}).Info("Delete stock API called")
	w.WriteHeader(http.StatusNoContent)
}
