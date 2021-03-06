package main

import (
	"encoding/json"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"github.com/lunny/html2md"
	"github.com/mitchellh/go-wordwrap"
	"html"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type FeedItem struct {
	Title   string
	Content string
}

func main() {
	item := FeedItem{}

	// Create client to fetch article from given url
	client := &http.Client{}
	url := fmt.Sprintf("https://mercury.postlight.com/parser?url=%s", os.Args[1])
	request, _ := http.NewRequest("GET", url, nil)

	// Append mercury postlight api key to request headers
	apikey := os.Getenv("MERCURY_API_KEY")
	request.Header.Set("x-api-key", apikey)
	response, _ := client.Do(request)
	defer response.Body.Close()

	// Fill the `item` attributes with api response
	err := json.NewDecoder(response.Body).Decode(&item)
	if err != nil {
		panic(err)
	}

	// Convert html to readable markdown
	md := html2md.Convert(item.Content)
	unescapedMd := html.UnescapeString(md)

	// Squash multiple lines blocks into single blank lines
	regex, _ := regexp.Compile(`(\s*\n){2,}`)
	output := regex.ReplaceAllString(unescapedMd, "\n\n")

	// Convert markdown wrappers to ANSI codes (to enhance subtitles)
	regex, _ = regexp.Compile(`\*\*(.*)\*\*`)
	output = regex.ReplaceAllString(output, fmt.Sprintf("%s", Bold("$1")))

	// Convert markdown wrappers to ANSI codes (to enhance subtitles)
	regex, _ = regexp.Compile("## (.*)")
	output = regex.ReplaceAllString(output, fmt.Sprintf("%s", Bold("$1")))

	// Remove leading whitespaces
	regex, _ = regexp.Compile(`[\n\n ][ \t]+`)
	output = regex.ReplaceAllString(output, "")

	// Wrap text to 80 columns to make the content more readable
	output = wordwrap.WrapString(output, 80)

	// Format article output with title and content
	output = fmt.Sprintf("%s\n%s", Bold(Red(item.Title)), output)
	cmd := exec.Command("/usr/bin/less", "-s")

	// Set `less` stdin to string Reader
	cmd.Stdin = strings.NewReader(output)

	// Set `less` stdout to os stdout
	cmd.Stdout = os.Stdout

	// Start the command and wait for user actions
	cmd.Start()
	cmd.Wait()
}
