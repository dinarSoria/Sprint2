package main

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"gopkg.in/gomail.v2"

	//"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/emailpriority"
	//"fyne.io/fyne"
	"fyne.io/fyne/v2"
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
	window := app.NewWindow("...")

	PrincipalLabel := widget.NewLabelWithStyle("Musmatch", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true, Monospace: true})

	nameEntry := widget.NewEntry()
	nameLabel := widget.NewLabel("Nombre:")

	lastnameEntry := widget.NewEntry()
	lastnameLabel := widget.NewLabel("Apellido:")

	emailEntry := widget.NewEntry()
	emailLabel := widget.NewLabel("Email:")

	genresEntry := widget.NewEntry()
	genresLabel := widget.NewLabel("Géneros musicales favoritos (separados por coma):")

	grid := container.NewGridWithColumns(2,
		container.NewPadded(nameLabel),
		container.NewPadded(nameEntry),
		container.NewPadded(lastnameLabel),
		container.NewPadded(lastnameEntry),
		container.NewPadded(emailLabel),
		container.NewPadded(emailEntry),
		container.NewPadded(genresLabel),
		container.NewPadded(genresEntry),
	)
	//grid.SetColumnAlignment(1, container.GridAlignTrailing)

	msLabel := widget.NewLabel("Ya existe un usuario registrado con ese email.")
	msLabel.Hide()

	resultsLabel := widget.NewLabel("Matchs:")
	resultsLabel.Hide()

	msLabel2 := widget.NewLabel("Email inválido")
	msLabel2.Hide()

	msLabel3 := widget.NewLabel("Por favor complete todos los campos.")
	msLabel3.Hide()

	searchButton := widget.NewButton("Buscar", func() {
		var p Person

		p.Musicalgenres = strings.Split(genresEntry.Text, ",")
		p.Email = emailEntry.Text
		filter := bson.M{
			"Musicalgenres": bson.M{"$in": p.Musicalgenres},
			"email":         bson.M{"$ne": p.Email},
		}

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
		resultsLabel.Show()

		// set up email message
		message := gomail.NewMessage()
		message.SetHeader("From", "soriadinar93@gmail.com")
		message.SetHeader("To", p.Email)
		message.SetHeader("Subject", "Resultados de la búsqueda de géneros musicales")

		// create the email body
		var body string
		body += "Estimado/a " + p.Name + ",\n\n"
		body += "Aquí están los resultados de la búsqueda de géneros musicales:\n\n"
		body += resultString
		message.SetBody("text/plain", body)

		// create the email sender
		sender := gomail.NewDialer("smtp.gmail.com", 587, "sprint2netlabs@gmail.com", "ekxgvfggqgjaiehh")

		if p.Email == "" {
			fmt.Println("La dirección de correo electrónico está vacía.")
			return
		}

		// send the email
		if err := sender.DialAndSend(message); err != nil {
			fmt.Println(err)
			fmt.Println("no se mandó")
		}
		fmt.Println("Correo electrónico enviado a", emailEntry.Text)
	})
	searchButton.Hide()

	StartButton := container.NewPadded(widget.NewButton("Start", func() {
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

	}))

	content := container.NewVBox(
		PrincipalLabel,
		widget.NewSeparator(),
		grid,
		//container.NewHBox(nameLabel, nameEntry),
		//container.NewHBox(lastnameLabel, lastnameEntry),
		//container.NewHBox(genresLabel, genresEntry),
		//container.NewHBox(emailLabel, emailEntry),
		//widget.NewSeparator(),
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
