package image

// Import go's image libraries
import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"strconv"
	"time"

	"io"
	"log"
	"net/http"
	"os"

	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"gopkg.in/yaml.v2"
)

/*DownloadFile This is a function that downloads a file from a URL and saves it to a file.*/
func DownloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

type rank struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Color []int  `json:"color"`
}

// Ranks is the global rank struct.
var Ranks []rank

var (
	backbanner image.Image
	normalfont font.Face
	titlefont  font.Face
	cfg        config.Configs
	rankmap    map[int]rank
)

// Initialize is a function to initialize the image library.
func Initialize() {

	ranksFile, err := ioutil.ReadFile("static/ranks.yaml") // Read the ranks file
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(ranksFile, &Ranks) // Parse the ranks file into the rank struct
	if err != nil {
		println(5)
		log.Println(err)
	}

	backbanner, err = gg.LoadImage("./images/backgroundbanner.png")

	if err != nil {
		log.Println(err)
	}
	normalfont, err = gg.LoadFontFace("./static/fonts/firacode.ttf", 28)
	if err != nil {
		log.Println(err)
	}
	titlefont, err = gg.LoadFontFace("./static/fonts/firacode.ttf", 40)
	if err != nil {
		log.Println(err)
	}

	cfg = config.Config()

	rankmap = make(map[int]rank)
	for _, v := range Ranks {
		rankmap[v.ID] = v
	}
}

// Account is a function to create a user banner if needed then return the path to it
func Account(uid string) string {
	db, err := raven.OpenSession(cfg.Ravenhost, cfg.Ravenport, "users")
	if err != nil {
		log.Println(err)
	}
	dc := gg.NewContext(800, 300)
	userdetail, err := db.Get(uid)
	if err != nil {
		log.Println(err)
	}
	// Add the user's avatar to the image. get it from https://discordapp.com/api/users/uid/avatarhash.png
	// Check if the user's image is already in images/users
	if _, err := os.Stat("images/users/" + uid + ".png"); os.IsNotExist(err) {
		err = DownloadFile(userdetail.PFP, "./images/users/"+uid+".png")
		if err != nil {
			log.Println(err)
		}

	} else {
		go func() {
			// Wait 1 second before downloading the image
			time.Sleep(1 * time.Second)
			err = DownloadFile(userdetail.PFP, "./images/users/"+uid+".png")
			if err != nil {
				log.Println(err)
			}
		}()
	}
	if err != nil {
		log.Println(err)
	}
	userimg, err := gg.LoadImage("./images/users/" + uid + ".png")
	if err != nil {
		log.Println(err)
	}

	var rankname string
	urank := rankmap[userdetail.Rank]
	rankname = urank.Name
	dc.DrawImage(backbanner, 0, 0)
	if urank.Color[0] == -1 {
		borderimg, err := gg.LoadImage("./images/rankborders/" + rankname + ".png")
		if err != nil {
			log.Println(err)
		}
		dc.DrawImage(resize.Resize(220, 220, borderimg, resize.NearestNeighbor), 40, 40)
	} else {
		dc.SetRGB255(urank.Color[0], urank.Color[1], urank.Color[2])
		dc.DrawRectangle(40, 40, 220, 220)
		dc.Fill()
	}
	dc.DrawImage(resize.Resize(200, 200, userimg, resize.NearestNeighbor), 50, 50)
	dc.SetRGBA(0.7, 0.7, 0.7, 0.6)
	dc.DrawRectangle(300, 50, 450, 200)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	dc.SetFontFace(titlefont)
	dc.DrawString(userdetail.Username, 320, 100)
	dc.SetFontFace(normalfont)
	dc.DrawString(fmt.Sprintf("Bal: $%.2f", userdetail.Bal), 320, 140)
	dc.DrawString("Stocks: "+strconv.Itoa(userdetail.Stocks), 320, 170)
	dc.DrawString("Rank: "+rankname, 320, 200)
	if err != nil {
		log.Println(err)
	}
	f, err := os.OpenFile("images/banners/"+uid+".jpeg", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Println(err)
	}
	err = jpeg.Encode(f, dc.Image(), &jpeg.Options{Quality: 80})
	if err != nil {
		log.Println(err)
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
	}
	dc = nil
	return "images/banners/" + uid + ".jpeg"

}
