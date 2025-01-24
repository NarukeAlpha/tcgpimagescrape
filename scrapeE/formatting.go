package scrapeE

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)

type Response struct {
	Errors  []string `json:"errors"`
	Results []Result `json:"results"`
}

type Result struct {
	CategoryId          int    `json:"categoryId"`
	CatalogGroupId      int    `json:"catalogGroupId"`
	CategoryName        string `json:"categoryName"`
	DisplayName         string `json:"displayName"`
	UrlName             string `json:"urlName"`
	CategoryDescription string `json:"categoryDescription"`
	CategoryPageTitle   string `json:"categoryPageTitle"`
	MpCanSearch         bool   `json:"mpCanSearch"`
}

func ClearTerminal() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Println("Unable to clear terminal", err)
	}
}

func ErrorCheck(err error, errstrng string) {
	if err != nil {
		log.Panicln(errstrng, err)
	}
}
