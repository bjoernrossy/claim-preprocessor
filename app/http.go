package blackberry

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/mjibson/appstats"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var templates = template.Must(template.New("signups-index").ParseGlob("templates/*.html"))
var signUpsPerPage = 20

type TemplateData struct {
	LogoutURL  string
	User       string
	Error      string
	FormValues map[string]string
	Data       interface{}
}

type AzureResponse struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string
}

type Claim struct {
	Positive    bool     `json:"positive"`
	Object      string   `json:"object"`
	Adjectives  []string `json:"adjectives"`
	Translation string   `json:"translation"`
}

func getToken(c appengine.Context) (string, error) {
	id := "StylegeistClaimPreprocessor"
	secret := "wT9I5LyXmx59yuXDAqxu5/m/HT/9j4onDIOQk/BXs10="
	uri := "https://datamarket.accesscontrol.windows.net/v2/OAuth2-13"

	client := urlfetch.Client(c)
	resp, err := client.PostForm(uri, url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{id},
		"client_secret": []string{secret},
		"scope":         []string{"http://api.microsofttranslator.com"},
	})

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Returned HTTP code %v")
	}

	var azure AzureResponse
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err = json.Unmarshal(bodyBytes, &azure); err != nil {
		return "", err
	}
	return azure.AccessToken, nil
}

func translateClaim(c appengine.Context, claim string) (string, error) {
	tok, err := getToken(c)
	if err != nil {
		return "", err
	}

	trValues := url.Values{
		"text": []string{claim},
		"to":   []string{"en"},
	}

	uri := fmt.Sprintf("http://api.microsofttranslator.com/v2/Ajax.svc/Translate?%v", trValues.Encode())
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tok))
	req.Header.Set("Accept-Charset", "utf-8")
	client := urlfetch.Client(c)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	translation := string(bodyBytes)
	return translation[4 : len(translation)-1], nil
}

func computeClaim(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	claim, err := translateClaim(c, v.Get("claim"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// fmt.Fprintf(w, fmt.Sprintf("%v %v", v.Get("claim"), claim))

	words := strings.Split(claim, " ")

	var positive bool
	var object string
	adjectives := make([]string, 0, 10)
	for _, v := range words {
		v = strings.ToLower(v)
		spaced := fmt.Sprintf(" %v ", v)
		if strings.Contains(" hate dislike reject detest abhor refuse decline don't bullshit worst bad ugly cheap gross hideous nasty awful terrible dreadful poor dire abominable shitty horrible ", spaced) {
			positive = false
		} else if strings.Contains(" like love adore desire want fond of into prefer fancy relish seek need awesome great nice lovely best good pretty cute neat bonny dinky nifty fine beautiful beauty fresh stunning amazing bodacious groovy yummy appealing delightful lovesome adorable alluring attractive inviting magnetic pleasing exciting seductive sweet trendy stylish classy chic elegant ", spaced) {
			positive = true
		}

		if strings.Contains(" clothes shirt tie shorts socks gloves cap pullover sweater jeans bikini boots suit cloth scarf panties pants jacket bag shoe top cardigan dress yarn t-shirt vest coat underwear hoodie polo track suit bracelet necklace gown strap legging boxer tie blazer junpsuit romper swimsuit tankini lingerie pajama robe capri corset bra waistcoat trunks beach shorts bag trousers handbag dessous blouse chemise singlet panties bow tie bodysuit jumper tailcoat smock guayabera costume glasses cut smoking cape bathrobe belt hat shoes ", spaced) {
			object = v
		}

		if strings.Contains(" striped casual green turquois turquoise orange blue grey beige red yellow pink light wide heavy brown black long sleeve club bohemian maxi mini summer cotton silk winter autumn spring lace chiffon bodycon vintage modern v-neck retro morden draped collarless ladylike fish net abstract camouflage classic loose slim chic elegant cheap expensive fleece white short long wide tight gold silver fishnet ", spaced) {
			adjectives = append(adjectives, v)
		}
	}

	j := Claim{
		Positive:    positive,
		Object:      object,
		Adjectives:  adjectives,
		Translation: claim,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, _ := json.MarshalIndent(j, "", "    ")
	fmt.Fprintf(w, "%v", string(b))
}

func newClaim(c appengine.Context, w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	message := v.Get("message")
	if err := templates.ExecuteTemplate(w, "claim-new.html", message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func init() {
	m := pat.New()
	m.Get("/", appstats.NewHandler(computeClaim))
	m.Get("/new", appstats.NewHandler(newClaim))
	http.Handle("/", m)
}
