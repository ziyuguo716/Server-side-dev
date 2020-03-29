package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
}

type statusError struct {
	code int
	prob string
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s - %d", e.prob, e.code)
}

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	// // 	- Add an HTTP header to the response with the name
	//  `Access-Control-Allow-Origin` and a value of `*`. This will
	//  allow cross-origin AJAX requests to your server.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// //    - Get the `url` query string parameter value from the request.
	// 	 If not supplied, respond with an http.StatusBadRequest error.
	URL := r.FormValue("url")
	if URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// //    - Call fetchHTML() to fetch the requested URL. See comments in that
	// 	 function for more details.
	bodyStream, err := fetchHTML(URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// //    - Call extractSummary() to extract the page summary meta-data,
	// 	 as directed in the assignment. See comments in that function
	// 	 for more details
	newPageSummary, err := extractSummary(URL, bodyStream)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// //   - Close the response HTML stream so that you dont leak resources.
	defer bodyStream.Close()
	// //    - Finally, respond with a JSON-encoded version of the PageSummary
	// 	 struct. That way the client can easily parse the JSON back into
	// 	 an object. Remember to tell the client that the response content
	// 	 type is JSON.
	json.NewEncoder(w).Encode(newPageSummary)

}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	/*TODO: Do an HTTP GET for the page URL. If the response status
	code is >= 400, return a nil stream and an error. If the response
	content type does not indicate that the content is a web page, return
	a nil stream and an error. Otherwise return the response body and
	no (nil) error.
	*/

	// Get http response
	res, err := http.Get(pageURL)
	if err != nil {
		return nil, errors.New("ERROR: Unable to get html")
	}
	if res.StatusCode >= 400 {
		return nil, &statusError{res.StatusCode, "ERROR: Status Code"}
	}

	//check content type is html
	header := res.Header.Get("Content-Type")
	if !strings.HasPrefix(header, "text/html") {
		return nil, errors.New("ERROR: This is not an html page")
	}

	return res.Body, nil
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	/*TODO: tokenize the `htmlStream` and extract the page summary meta-data
	according to the assignment description.
	*/

	//initialize a new instance of PageSummary
	newPageSummary := new(PageSummary)
	tokenizer := html.NewTokenizer(htmlStream)

	//loop over all tokens until there's no more
	for tokenType := tokenizer.Next(); tokenType != html.ErrorToken; tokenType = tokenizer.Next() {
		//process the token according to the token type...
		//if this is a start tag token or self closing token
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()

			//Check the tag name
			if "link" == token.Data {
				icon := new(PreviewImage)
				for _, attr := range token.Attr {
					name := attr.Key
					val := attr.Val
					// Check the attribute name
					switch name {
					case "rel":
						if val != "icon" {
							break
						}
					case "href":
						// URL := val
						// //check if URL is absolute path or relative path
						// if strings.HasPrefix(URL, "http:") {
						// 	icon.URL = URL
						// } else {
						// 	relativeIndex := strings.LastIndex(pageURL, ".com") + 3
						// 	fullURL := pageURL[:relativeIndex+1] + URL
						// 	icon.URL = fullURL
						// }
						URL, err := relURLToAbsURL(pageURL, val)
						if err != nil {
							return nil, err
						}
						icon.URL = URL
					case "sizes":
						size := val
						//split the height and width in sizes
						splitIndex := strings.IndexAny(size, "x")
						if splitIndex > 0 {
							h, _ := strconv.Atoi(size[:splitIndex])
							w, _ := strconv.Atoi(size[splitIndex+1:])
							icon.Height = h
							icon.Width = w
						}
					case "type":
						icon.Type = val
					}
				}
				newPageSummary.Icon = icon
			}
			if "meta" == token.Data {
				property, content := "", ""
				for _, attr := range token.Attr {
					name := attr.Key
					val := attr.Val
					if name == "property" {
						valString := val
						colonIdx := strings.Index(valString, ":")
						//get rid of the og: in property
						property = valString[colonIdx+1:]
					} else if name == "name" {
						property = val
					} else if name == "content" {
						content = val
					}

					//only if we have crawled both property and content value
					if property != "" && content != "" {
						switch property {
						case "author":
							newPageSummary.Author = content
						case "title":
							if newPageSummary.Title == "" {
								newPageSummary.Title = content
							}
						case "type":
							newPageSummary.Type = content
						case "url":
							newPageSummary.URL = content
						case "keywords":
							words := strings.Split(content, ",")
							for i := range words {
								words[i] = strings.TrimSpace(words[i])
							}
							newPageSummary.Keywords = words
						case "site_name":
							newPageSummary.SiteName = content
						case "description":
							if newPageSummary.Description == "" {
								newPageSummary.Description = content
							}
						case "image":
							newImage := new(PreviewImage)
							URL, err := relURLToAbsURL(pageURL, content)
							if err != nil {
								return nil, err
							}
							newImage.URL = URL
							newPageSummary.AddImage(newImage)
						case "image:secure_url":
							sURL, err := relURLToAbsURL(pageURL, content)
							if err != nil {
								return nil, err
							}
							newPageSummary.GetLast().SecureURL = sURL
						case "image:width":
							w, _ := strconv.Atoi(content)
							newPageSummary.GetLast().Width = w
						case "image:height":
							h, _ := strconv.Atoi(content)
							newPageSummary.GetLast().Height = h
						case "image:alt":
							newPageSummary.GetLast().Alt = content
						case "image:type":
							newPageSummary.GetLast().Type = content
						}
						//reset them to empty string for the next attribute
						property = ""
						content = ""
					}

				}
			}
			if "title" == token.Data {
				//Ensure meta tag overrides title
				if newPageSummary.Title == "" {
					tokenType = tokenizer.Next()
					if tokenType == html.TextToken {
						newPageSummary.Title = tokenizer.Token().Data
					}
				}
			}
		}
	}
	return newPageSummary, nil
}

//AddImage ():for PageSummary struct to add new image
func (page *PageSummary) AddImage(img *PreviewImage) {
	page.Images = append(page.Images, img)
}

//GetLast (): get last image from array
func (page *PageSummary) GetLast() *PreviewImage {
	return page.Images[len(page.Images)-1]
}

func relURLToAbsURL(baseURL string, relURL string) (string, error) {
	parsedRelURL, err := url.Parse(relURL)
	if err != nil {
		return "", err
	}
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	absURL := parsedBaseURL.ResolveReference(parsedRelURL)
	return absURL.String(), nil
}
