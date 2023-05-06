package main

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	//"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/emailpriority"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	//"golang.org/x/text/search"

	//"fyne.io/fyne/v2/internal/widget"
	"fyne.io/fyne/v2/widget"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	Name          string   `bson:"name"`
	Lastname      string   `bson:"lastname"`
	Email         string   `bson:"email"`
	Musicalgenres []string `bson:"Musicalgenres"`
}

func main() {
	// Set the URI of your MongoDB Atlas cluster
	uri := "mongodb+srv://SdinarNetlabs:123dinar@cluster0.2ss0vbw.mongodb.net/?retryWrites=true&w=majority"

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB Atlas
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB Atlas!")

	DataB := client.Database("DataB")
	Collection := DataB.Collection("Persons")

	app := app.New()
	window := app.NewWindow("Menu")

	nameEntry := widget.NewEntry()
	nameLabel := widget.NewLabel("Nombre:")

	lastnameEntry := widget.NewEntry()
	lastnameLabel := widget.NewLabel("Apellido:")

	emailEntry := widget.NewEntry()
	emailLabel := widget.NewLabel("Email:")

	genresEntry := widget.NewEntry()
	genresLabel := widget.NewLabel("Géneros musicales favoritos (separados por coma):")

	msLabel := widget.NewLabel("Ya existe un usuario registrado con ese email.")
	msLabel.Hide()

	resultsLabel := widget.NewLabel("Results")

	msLabel2 := widget.NewLabel("Email inválido")
	msLabel2.Hide()

	msLabel3 := widget.NewLabel("Por favor complete todos los campos.")
	msLabel3.Hide()

	searchButton := widget.NewButton("Buscar", func() {
		var p Person

		p.Musicalgenres = strings.Split(genresEntry.Text, ",")
		filter := bson.M{"Musicalgenres": bson.M{"$in": p.Musicalgenres}}

		cur, err := Collection.Find(context.Background(), filter)
		if err != nil {
			panic(err)
		}
		defer cur.Close(context.Background())

		var results []Person
		for cur.Next(context.Background()) {
			var result Person
			err := cur.Decode(&result)
			if err != nil {
				panic(err)
			}
			results = append(results, result)
		}

		if len(results) == 0 {
			fmt.Println("No se encontraron resultados.")
			return
		}

		var resultString string
		resultString += "Resultados:\n"
		for _, r := range results {
			resultString += fmt.Sprintf("%s %s - %s\n", r.Name, r.Lastname, r.Email)
		}

		resultsLabel.SetText(resultString)
	})
	searchButton.Hide()

	StartButton := widget.NewButton("Start", func() {
		if nameEntry.Text == "" || lastnameEntry.Text == "" || emailEntry.Text == "" || genresEntry.Text == "" {
			msLabel3.Show()
			return
		}
		var p Person

		p.Name = nameEntry.Text
		p.Lastname = lastnameEntry.Text
		// Check if the email is valid
		email, err := mail.ParseAddress(emailEntry.Text)
		if err != nil {
			msLabel2.Show()
			return
		}
		p.Email = email.Address
		p.Musicalgenres = strings.Split(genresEntry.Text, ",")

		// Check if a person with the same email already exists
		filter := bson.M{"email": p.Email}
		existingPerson := Collection.FindOne(context.Background(), filter)
		var per Person
		if err := existingPerson.Decode(&per); err == nil {
			// An object with the same email already exists, inform the user
			msLabel.Show()
			return
		}

		result, err := Collection.InsertOne(context.Background(), p)
		if err != nil {
			panic(err)
		}
		searchButton.Show()
		fmt.Println("Inserted document with ID:", result.InsertedID)
		widget.NewLabel(fmt.Sprintf("Usuario registrado con ID %v", result.InsertedID)).Show()

	})

	content := container.NewVBox(
		container.NewHBox(nameLabel, nameEntry),
		container.NewHBox(lastnameLabel, lastnameEntry),
		container.NewHBox(genresLabel, genresEntry),
		container.NewHBox(emailLabel, emailEntry),
		StartButton,
		msLabel,
		msLabel2,
		msLabel3,
		searchButton,
		resultsLabel,
	)

	window.SetContent(content)
	window.ShowAndRun()
}
