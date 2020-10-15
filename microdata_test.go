// Copyright 2015 Lars Wiegman. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

package microdata

import (
	"bytes"
	"encoding/json"
	"github.com/bradleyjkemp/cupaloy"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseItemScope(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Properties)
	expected := 1
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
	}
}

func TestParseItemType(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Types[0]
	expected := "http://example.com/Person"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemRef(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Movie" itemref="properties">
			<p><span itemprop="name">Rear Window</span> is a movie from 1954.</p>
		</div>
		<ul id="properties">
			<li itemprop="genre">Thriller</li>
			<li itemprop="description">A homebound photographer spies on his neighbours.</li>
		</ul>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"genre", "Thriller"},
		{"description", "A homebound photographer spies on his neighbours."},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseItemProp(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseItemId(t *testing.T) {
	html := `
		<ul itemscope itemtype="http://example.com/Book" itemid="urn:isbn:978-0141196404">
			<li itemprop="title">The Black Cloud</li>
			<li itemprop="author">Fred Hoyle</li>
		</ul>`

	data := ParseData(html, t)

	result := data.Items[0].ID
	expected := "urn:isbn:978-0141196404"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseHref(t *testing.T) {
	html := `
		<html itemscope itemtype="http://example.com/Person">
			<head>
				<link itemprop="linkTest" href="http://example.com/cde">
			<head>
			<div>
				<a itemprop="aTest" href="http://example.com/abc" /></a>
				<area itemprop="areaTest" href="http://example.com/bcd" />
			</div>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"aTest", "http://example.com/abc"},
		{"areaTest", "http://example.com/bcd"},
		{"linkTest", "http://example.com/cde"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseSrc(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Videocast">
			<audio itemprop="audioTest" src="http://example.com/abc" />
			<embed itemprop="embedTest" src="http://example.com/bcd" />
			<iframe itemprop="iframeTest" src="http://example.com/cde"></iframe>
			<img itemprop="imgTest" src="http://example.com/def" />
			<source itemprop="sourceTest" src="http://example.com/efg" />
			<track itemprop="trackTest" src="http://example.com/fgh" />
			<video itemprop="videoTest" src="http://example.com/ghi" />
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"audioTest", "http://example.com/abc"},
		{"embedTest", "http://example.com/bcd"},
		{"iframeTest", "http://example.com/cde"},
		{"imgTest", "http://example.com/def"},
		{"sourceTest", "http://example.com/efg"},
		{"trackTest", "http://example.com/fgh"},
		{"videoTest", "http://example.com/ghi"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseMetaContent(t *testing.T) {
	html := `
		<html itemscope itemtype="http://example.com/Person">
			<meta itemprop="length" content="1.70" />
		</html>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["length"][0].(string)
	expected := "1.70"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseValue(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Container">
			<data itemprop="capacity" value="80">80 liters</data>
			<meter itemprop="volume" min="0" max="100" value="25">25%</meter>
		</div>`

	var testTable = []struct {
		propName string
		expected string
	}{
		{"capacity", "80"},
		{"volume", "25"},
	}

	data := ParseData(html, t)

	for _, test := range testTable {
		if result := data.Items[0].Properties[test.propName][0].(string); result != test.expected {
			t.Errorf("Result should have been \"%s\", but it was \"%s\"", test.expected, result)
		}
	}
}

func TestParseDatetime(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<time itemprop="birthDate" datetime="1993-10-02">22 years</time>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["birthDate"][0].(string)
	expected := "1993-10-02"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseText(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Product">
			<span itemprop="price">3.95</span>
		</div>`

	data := ParseData(html, t)

	result := data.Items[0].Properties["price"][0].(string)
	expected := "3.95"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseMultiItemTypes(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Park http://example.com/Zoo">
			<span itemprop="name">ZooParc Overloon</span>
		</div>`

	data := ParseData(html, t)

	result := len(data.Items[0].Types)
	expected := 2
	if result != expected {
		t.Errorf("Result should have been \"%d\", but it was \"%d\"", expected, result)
	}
}

func TestJSON(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
			<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseHTML(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	u, _ := url.Parse("http://blog.example.com/progress-report")
	contentType := "charset=utf-8"

	_, result := ParseHTML(buf, contentType, u)
	if result != nil {
		t.Errorf("Result should have been nil, but it was \"%s\"", result)
	}
}

func TestParseURL(t *testing.T) {
	html := `
		<div itemscope itemtype="http://example.com/Person">
			<p>My name is <span itemprop="name">Penelope</span>.</p>
		</div>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer ts.Close()

	data, err := ParseURL(ts.URL)
	if err != nil {
		t.Error(err)
	}

	result := data.Items[0].Properties["name"][0].(string)
	expected := "Penelope"
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseURLSnapshot(t *testing.T) {
	data, err := ParseURL("https://www.idnes.cz/zlin/zpravy/smrtelna-nehoda-motorkar-srazil-zenu.A150719_165426_zlin-zpravy_kol")
	if err != nil {
		t.Error(err)
	}

	cupaloy.SnapshotT(t, data)
}

func TestNestedItems(t *testing.T) {
	html := `
		<div>
			<div itemscope itemtype="http://example.com/Person">
				<p>My name is <span itemprop="name">Penelope</span>.</p>
				<p>I am <date itemprop="age" value="22">22 years old.</span>.</p>
				<div itemscope itemtype="http://example.com/Breadcrumb">
					<a itemprop="url" href="http://example.com/users/1"><span itemprop="title">profile</span></a>
				</div>
			</div>
		</div>`

	data := ParseData(html, t)

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://example.com/Person"],"properties":{"age":["22 years old.."],"name":["Penelope"]}},{"type":["http://example.com/Breadcrumb"],"properties":{"title":["profile"],"url":["http://example.com/users/1"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func ParseData(html string, t *testing.T) *Microdata {
	r := strings.NewReader(html)
	u, _ := url.Parse("http://example.com")

	p, err := newParser(r, "utf-8", u)
	if err != nil {
		t.Error(err)
	}

	data, err := p.parse()
	if err != nil {
		t.Error(err)
	}
	return data
}

func TestParseW3CBookSnippet(t *testing.T) {
	buf := bytes.NewBufferString(bookSnippet)
	u, _ := url.Parse("")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://vocab.example.net/book"],"properties":{"author":["Peter F. Hamilton"],"pubdate":["1996-01-26"],"title":["The Reality Dysfunction"]},"id":"urn:isbn:0-330-34032-8"}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CGalleySnippet(t *testing.T) {
	buf := bytes.NewBufferString(gallerySnippet)
	u, _ := url.Parse("")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://n.whatwg.org/work"],"properties":{"license":["http://www.opensource.org/licenses/mit-license.php"],"title":["The house I found."],"work":["/images/house.jpeg"]}},{"type":["http://n.whatwg.org/work"],"properties":{"license":["http://www.opensource.org/licenses/mit-license.php"],"title":["The mailbox."],"work":["/images/mailbox.jpeg"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseW3CBlogSnippet(t *testing.T) {
	buf := bytes.NewBufferString(blogSnippet)
	u, _ := url.Parse("http://blog.example.com/progress-report")
	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	result := string(b)
	expected := `{"items":[{"type":["http://schema.org/BlogPosting"],"properties":{"comment":[{"type":["http://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["http://schema.org/Person"],"properties":{"name":["Greg"]}}],"url":["http://blog.example.com/progress-report#c1"]}},{"type":["http://schema.org/UserComments"],"properties":{"commentTime":["2013-08-29"],"creator":[{"type":["http://schema.org/Person"],"properties":{"name":["Charlotte"]}}],"url":["http://blog.example.com/progress-report#c2"]}}],"datePublished":["2013-08-29"],"headline":["Progress report"],"url":["http://blog.example.com/progress-report?comments=0"]}}]}`
	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \"%s\"", expected, result)
	}
}

func TestParseStackOverflowSnippet(t *testing.T) {
	buf := bytes.NewBufferString(stackOverflowSnippet)
	u, err := url.Parse("http://blog.example.com/progress-report")
	if err != nil {
		t.Error(err)
	}

	data, err := ParseHTML(buf, "charset=utf-8", u)
	if err != nil {
		t.Error(err)
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	result := string(b)
	expected := `{"items":[{"type":["https://schema.org/WebPage"],"properties":{"breadcrumb":["Aktien»Nachrichten»GENERIC GOLD AKTIE»Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million"],"image":["https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-1jahrklein-frankfurt.png"]}},{"type":["http://schema.org/Product"],"properties":{"image":["https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-1jahrklein-frankfurt.png"],"name":["GENERIC GOLD CORP"],"offers":[{"type":["http://schema.org/Offer"],"properties":{"price":["0.28"],"priceCurrency":["EUR"]}}],"priceValidUntil":["2020-10-15T09:29:28.0000000"],"productID":["wkn:A2JAE9","wkn:A2JAE9"],"seller":["Frankfurt"],"url":["https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm"]}},{"type":["http://schema.org/Article"],"properties":{"aggregateRating":[{"type":["http://schema.org/AggregateRating"],"properties":{"bestRating":["5"],"itemReviewed":["Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million"],"ratingCount":["2"],"ratingValue":["4,5"],"worstRating":["1"]}}],"articleBody":["Toronto, Ontario--(Newsfile Corp. - July 27, 2020) - Generic Gold Corp. (CSE: GGC) (FSE: 1WD) (\"Generic Gold\" or the \"Company\") is pleased to announce, further to its press release of July 16, 2020, the upsize of its fully subscribed \"best efforts\" private placement offering, led by StephenAvenue Securities Inc. (the \"Agent\") as sole agent and sole bookrunner (the \"Offering\"), for aggregate gross proceeds of up to $7,000,000, through the issuance of units (each, a \"Unit\") at a price of $0.35 per Unit and flow-through units (each, a \"FT Unit\") at a price of $0.40 per FT Unit (together, the Units and the FT Units, the \"Offered Securities\").The net proceeds from the sale of the Units will be used for general working capital and exploration purposes. The gross proceeds from the sale of the FT Units will be used by the Company to incur eligible \"Canadian exploration expenses\" that will qualify as \"flow-through mining expenditures\" (as such terms are defined in the Income Tax Act (Canada)) (the \"Qualifying Expenditures\") related to the Company's projects in Canada. All Qualifying Expenditures will be renounced in favour of the subscribers of the FT Units effective December 31, 2020. It is anticipated that most of the funds derived from the sale of the FT Units will be used to explore the Company's recently acquired Belvais project which is contiguous to Amex Exploration Inc.  (refer to the Company's press release of July 7, 2020).The Offering is expected to close on or about August 6, 2020 (the \"Closing Date\"), or such other date as agreed between the Company and the Agent. The completion of the Offering is subject to certain closing conditions including, but not limited to, the receipt of all necessary regulatory and other approvals including the approval of the Canadian Securities Exchange. All Offered Securities will be subject to a statutory hold period of four months and one day from the Closing Date.The Offered Securities have not been and will not be registered under the U.S. Securities Act of 1933, as amended, and may not be offered or sold in the United States absent registration or an applicable exemption from the registration requirements. This press release shall not constitute an offer to sell or the solicitation of an offer to buy nor shall there be any sale of the Offered Securities in any State in which such offer, solicitation or sale would be unlawful.About Generic GoldGeneric Gold is a Canadian mineral exploration company focused on gold projects in the Abitibi Greenstone Belt in Quebec, Canada and Tintina Gold Belt in the Yukon Territory of Canada. The Company's Quebec exploration portfolio consists of three properties covering 8,148 hectares proximal to the town of Normétal and Amex Exploration's Perron project. The Company's Yukon exploration portfolio consists of several projects with a total land position of greater than 35,000 hectares, all of which are 100% owned by Generic Gold. Several of these projects are in close proximity to significant gold projects, including Goldcorp's Coffee project, Victoria Gold's Eagle Gold project, White Gold's Golden Saddle project, and Western Copper \u0026 Gold's Casino project. For information on the Company's property portfolio, visit the Company's website at genericgold.ca.For further information contact: Generic Gold Corp. Richard Patricio, President and CEO Tel: 416-456-6529 rpatricio@genericgold.caStephenAvenue Securities Inc.Daniel CappuccittiTel: 416-479-4478ecm@stephenavenue.comNEITHER THE CANADIAN SECURITIES EXCHANGE NOR THEIR REGULATION SERVICES PROVIDERS ACCEPT RESPONSIBILITY FOR THE ADEQUACY OR ACCURACY OF THIS RELEASE.Certain statements in this press release are \"forward-looking\" statements within the meaning of Canadian securities legislation. All statements, other than statements of historical fact, included herein are forward-looking information.  Forward-looking statements are necessarily based upon the current belief, opinions and expectations of management that, while considered reasonable by the Company, are inherently subject to business, economic, competitive, political and social uncertainties and other contingencies. Many factors could cause the Company's actual results to differ materially from those expressed or implied in the forward-looking statements. Accordingly, readers should not place undue reliance on forward-looking statements and forward-looking information. The Company does not undertake to update any forward-looking statements or forward-looking information that are incorporated by reference herein, except in accordance with applicable securities laws. Investors are cautioned not to put undue reliance on forward-looking statements due to the inherent uncertainty therein. We seek safe harbour.To view the source version of this press release, please visit https://www.newsfilecorp.com/release/60535GENERIC GOLD-Aktie komplett kostenlos handeln - auf Smartbroker.de"],"author":["Newsfile"],"datePublished":["2020-07-27T14:22"],"headline":["Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million"],"interactionCount":["UserPageVisits:305"],"keywords":["Generic Gold Announces Upsizing Fully Subscribed Private Placement Million"],"publisher":["https://www.finanznachrichten.de"],"url":["https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"]}}]}`

	if result != expected {
		t.Errorf("Result should have been \"%s\", but it was \n \"%s\"", expected, result)
	}
}

func BenchmarkParser(b *testing.B) {
	buf := bytes.NewBufferString(blogSnippet)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, _ := url.Parse("http://blog.example.com/progress-report")
		_, err := ParseHTML(buf, "utf-8", u)
		if err != nil && err != io.EOF {
			b.Error(err)
		}
	}
}

// This HTML snippet is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var bookSnippet string = `
<dl itemscope
    itemtype="http://vocab.example.net/book"
    itemid="urn:isbn:0-330-34032-8">
 <dt>Title</td>
 <dd itemprop="title">The Reality Dysfunction</dd>
 <dt>Author</dt>
 <dd itemprop="author">Peter F. Hamilton</dd>
 <dt>Publication date</dt>
 <dd><time itemprop="pubdate" datetime="1996-01-26">26 January 1996</time></dd>
</dl>`

// This HTML snippet is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var gallerySnippet string = `
<!DOCTYPE HTML>
<html>
 <head>
  <title>Photo gallery</title>
 </head>
 <body>
  <h1>My photos</h1>
  <figure itemscope itemtype="http://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/house.jpeg" alt="A white house, boarded up, sits in a forest.">
   <figcaption itemprop="title">The house I found.</figcaption>
  </figure>
  <figure itemscope itemtype="http://n.whatwg.org/work" itemref="licenses">
   <img itemprop="work" src="images/mailbox.jpeg" alt="Outside the house is a mailbox. It has a leaflet inside.">
   <figcaption itemprop="title">The mailbox.</figcaption>
  </figure>
  <footer>
   <p id="licenses">All images licensed under the <a itemprop="license"
   href="http://www.opensource.org/licenses/mit-license.php">MIT
   license</a>.</p>
  </footer>
 </body>
</html>`

// This HTML document is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var blogSnippet string = `
<!DOCTYPE HTML>
<title>My Blog</title>
<article itemscope itemtype="http://schema.org/BlogPosting">
 <header>
  <h1 itemprop="headline">Progress report</h1>
  <p><time itemprop="datePublished" datetime="2013-08-29">today</time></p>
  <link itemprop="url" href="?comments=0">
 </header>
 <p>All in all, he's doing well with his swim lessons. The biggest thing was he had trouble
 putting his head in, but we got it down.</p>
 <section>
  <h1>Comments</h1>
  <article itemprop="comment" itemscope itemtype="http://schema.org/UserComments" id="c1">
   <link itemprop="url" href="#c1">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="http://schema.org/Person">
     <span itemprop="name">Greg</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">15 minutes ago</time></p>
   </footer>
   <p>Ha!</p>
  </article>
  <article itemprop="comment" itemscope itemtype="http://schema.org/UserComments" id="c2">
   <link itemprop="url" href="#c2">
   <footer>
    <p>Posted by: <span itemprop="creator" itemscope itemtype="http://schema.org/Person">
     <span itemprop="name">Charlotte</span>
    </span></p>
    <p><time itemprop="commentTime" datetime="2013-08-29">5 minutes ago</time></p>
   </footer>
   <p>When you say "we got it down"...</p>
  </article>
 </section>
</article>`

// This HTML snippet is taken from the W3C Working Group website at http://www.w3.org/TR/microdata.
var stackOverflowSnippet = `
	<!DOCTYPE html><html lang="de"><head><title>Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million</title><link rel="dns-prefetch" href="//fns1.de" /><meta http-equiv="Content-Type" content="text/html; charset=utf-8" /><meta name="keywords" content="Generic, Gold, Announces, Upsizing, Fully, Subscribed, Private, Placement, Million" /><meta name="description" content="Toronto, Ontario--(Newsfile Corp. - July 27, 2020) - Generic Gold Corp. (CSE: GGC) (FSE: 1WD) (&quot;Generic Gold&quot; or the &quot;Company&quot;) is pleased to announce, further to its press release of July 16, 2020, the" /><meta name="robots" content="noodp,index,follow" /><meta property="og:title" content="Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million" /><meta property="og:url" content="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm" /><meta property="og:site_name" content="FinanzNachrichten.de" /><meta property="og:image" content="https://fns1.de/g/fb.png" /><meta property="og:type" content="article" /><meta property="og:description" content="Toronto, Ontario--(Newsfile Corp. - July 27, 2020) - Generic Gold Corp. (CSE: GGC) (FSE: 1WD) (&quot;Generic Gold&quot; or the &quot;Company&quot;) is pleased to announce, further to its press release of July 16, 2020, the" /><meta property="fb:admins" content="100001851463444" /><link rel="canonical" href="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm" /><meta name="apple-itunes-app" content="app-id=576714727" /><meta name="google-play-app" content="app-id=de.finanznachrichten.app"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="shortcut icon" href="https://fns1.de/g/favicon.ico" /><link rel="search" type="application/opensearchdescription+xml" title="FinanzNachrichten.de Suche" href="https://fns1.de/suche/fn-osd-2.xml" /><link href="http://www.finanznachrichten.de/rss-aktien-nachrichten/" title="Aktuelle Nachrichten" type="application/rss+xml" rel="alternate" /><link href="http://www.finanznachrichten.de/rss-aktien-analysen/" title="Aktienanalysen" type="application/rss+xml" rel="alternate" /><link href="http://www.finanznachrichten.de/rss-aktien-adhoc/" title="Ad hoc-Mitteilungen" type="application/rss+xml" rel="alternate" /><link href="http://www.finanznachrichten.de/rss-news/" title="Englischsprachige Nachrichten" type="application/rss+xml" rel="alternate" /><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootswatch/3.4.0/yeti/bootstrap.min.css" integrity="sha256-ICOIXkZvfEjsPIVGgvAVTRNsYQbRELh04MoGaItVyuw=" crossorigin="anonymous" /><link href="https://fns1.de/css/fn216.css" rel="stylesheet" /><!-- Global site tag (gtag.js) - Google Analytics --><script async src="https://www.googletagmanager.com/gtag/js?id=UA-55465-3"></script><script>window.dataLayer = window.dataLayer || []; function gtag() { dataLayer.push(arguments); } gtag('js', new Date()); gtag('config', 'UA-55465-3', { 'anonymize_ip': true });</script><script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.min.js" integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo=" crossorigin="anonymous"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/jquery-migrate/3.0.1/jquery-migrate.min.js" integrity="sha256-F0O1TmEa4I8N24nY0bya59eP6svWcshqX1uzwaWC4F4=" crossorigin="anonymous"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/OwlCarousel2/2.3.4/owl.carousel.min.js" integrity="sha256-pTxD+DSzIwmwhOqTFN+DB+nHjO4iAsbgfyFq5K5bcE0=" crossorigin="anonymous"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/3.4.0/js/bootstrap.min.js" integrity="sha256-kJrlY+s09+QoWjpkOrXXwhxeaoDz9FW5SaxF8I0DibQ=" crossorigin="anonymous"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/moment.js/2.26.0/moment-with-locales.min.js" integrity="sha256-4HOrwHz9ACPZBxAav7mYYlbeMiAL0h6+lZ36cLNpR+E=" crossorigin="anonymous"></script><!-- Quantcast Choice. Consent Manager Tag v2.0 (for TCF 2.0) --><script type="text/javascript" async=true>(function () {
                var host = window.location.hostname;
                var element = document.createElement('script');
                var firstScript = document.getElementsByTagName('script')[0];
                var url = 'https://quantcast.mgr.consensu.org'
                    .concat('/choice/', 'XQPkmN_wVNAmT', '/', host, '/choice.js')
                var uspTries = 0;
                var uspTriesLimit = 3;
                element.async = true;
                element.type = 'text/javascript';
                element.src = url;

                firstScript.parentNode.insertBefore(element, firstScript);

                function makeStub() {
                    var TCF_LOCATOR_NAME = '__tcfapiLocator';
                    var queue = [];
                    var win = window;
                    var cmpFrame;

                    function addFrame() {
                        var doc = win.document;
                        var otherCMP = !!(win.frames[TCF_LOCATOR_NAME]);

                        if (!otherCMP) {
                            if (doc.body) {
                                var iframe = doc.createElement('iframe');

                                iframe.style.cssText = 'display:none';
                                iframe.name = TCF_LOCATOR_NAME;
                                doc.body.appendChild(iframe);
                            } else {
                                setTimeout(addFrame, 5);
                            }
                        }
                        return !otherCMP;
                    }

                    function tcfAPIHandler() {
                        var gdprApplies;
                        var args = arguments;

                        if (!args.length) {
                            return queue;
                        } else if (args[0] === 'setGdprApplies') {
                            if (
                                args.length > 3 &&
                                args[2] === 2 &&
                                typeof args[3] === 'boolean'
                            ) {
                                gdprApplies = args[3];
                                if (typeof args[2] === 'function') {
                                    args[2]('set', true);
                                }
                            }
                        } else if (args[0] === 'ping') {
                            var retr = {
                                gdprApplies: gdprApplies,
                                cmpLoaded: false,
                                cmpStatus: 'stub'
                            };

                            if (typeof args[2] === 'function') {
                                args[2](retr);
                            }
                        } else {
                            queue.push(args);
                        }
                    }

                    function postMessageEventHandler(event) {
                        var msgIsString = typeof event.data === 'string';
                        var json = {};

                        try {
                            if (msgIsString) {
                                json = JSON.parse(event.data);
                            } else {
                                json = event.data;
                            }
                        } catch (ignore) { }

                        var payload = json.__tcfapiCall;

                        if (payload) {
                            window.__tcfapi(
                                payload.command,
                                payload.version,
                                function (retValue, success) {
                                    var returnMsg = {
                                        __tcfapiReturn: {
                                            returnValue: retValue,
                                            success: success,
                                            callId: payload.callId
                                        }
                                    };
                                    if (msgIsString) {
                                        returnMsg = JSON.stringify(returnMsg);
                                    }
                                    event.source.postMessage(returnMsg, '*');
                                },
                                payload.parameter
                            );
                        }
                    }

                    while (win) {
                        try {
                            if (win.frames[TCF_LOCATOR_NAME]) {
                                cmpFrame = win;
                                break;
                            }
                        } catch (ignore) { }

                        if (win === window.top) {
                            break;
                        }
                        win = win.parent;
                    }
                    if (!cmpFrame) {
                        addFrame();
                        win.__tcfapi = tcfAPIHandler;
                        win.addEventListener('message', postMessageEventHandler, false);
                    }
                };

                if (typeof module !== 'undefined') {
                    module.exports = makeStub;
                } else {
                    makeStub();
                }

                var uspStubFunction = function () {
                    var arg = arguments;
                    if (typeof window.__uspapi !== uspStubFunction) {
                        setTimeout(function () {
                            if (typeof window.__uspapi !== 'undefined') {
                                window.__uspapi.apply(window.__uspapi, arg);
                            }
                        }, 500);
                    }
                };

                var checkIfUspIsReady = function () {
                    uspTries++;
                    if (window.__uspapi === uspStubFunction && uspTries < uspTriesLimit) {
                        console.warn('USP is not accessible');
                    } else {
                        clearInterval(uspInterval);
                    }
                };

                if (typeof window.__uspapi === 'undefined') {
                    window.__uspapi = uspStubFunction;
                    var uspInterval = setInterval(checkIfUspIsReady, 6000);
                }
            })();</script><!-- End Quantcast Choice. Consent Manager Tag v2.0 (for TCF 2.0) --><script async src="https://cdn.insurads.com/bootstrap/JZTPZVBW.js"></script><meta property="og:image" content="https://orders.newsfilecorp.com/files/3923/60535_1df1c2fe4451d8fe_logo.jpg" /><meta property="og:image" content="https://www.newsfilecorp.com/newsinfo/60535/207" /><meta property="og:image" content="https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-1jahrklein-frankfurt.png" /><meta property="og:image" content="https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-intraklein-frankfurt.png" /><meta property="og:image" content="https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-intraday-frankfurt.png" /><script>(function () {
                var s = document.createElement('script');
                s.type = 'text/javascript';
                s.async = true;
                s.src = '//d.finative.eu/d/init';
                var sc = document.getElementsByTagName('script')[0];
                sc.parentNode.insertBefore(s, sc);
            })();</script><script>window.impactifyTag = window.impactifyTag || [];
                impactifyTag.push({
                    "appId": "finanznachrichten.de",
                    "format": "screen",
                    "style": "impact",
                    "onNoAd": function () { }
                });
                (function (d, s, id) {
                    var js, ijs = d.getElementsByTagName(s)[0];
                    if (d.getElementById(id)) return;
                    js = d.createElement(s); js.id = id;
                    js.src = 'https://ad.impactify.io/static/ad/tag.js';
                    ijs.parentNode.insertBefore(js, ijs);
                }(document, 'script', 'impactify-sdk'));</script></head><body id="artikel_index" itemscope itemtype="https://schema.org/WebPage" class="l-xl"><div id="fb-root"></div><div id="mantel"><aside id="slide-out"></aside><div id="sideteaser"><div class="anzeige">Anzeige</div><div class="close">&#10060;</div><div class="headline"></div><div class="text"></div><img class="image" /><div class="more">Mehr »</div></div><div id="adspacer"></div><div id="inhalt" class="inner-wrapper small-header"><div id="fnk"><div id="fn-lf"></div><a id="btn-login" class="slide-out-open" href="https://www.finanznachrichten.de/watchlist/login.htm"><i class="icon fn-icon-torso"></i>Login</a><div class="fn-mobile-logo"><a title="Aktuelle Nachrichten zu Aktien, Börse und Finanzen" href="https://www.finanznachrichten.de/"><img src="https://fns1.de/img/logo.svg" alt="FinanzNachrichten.de" /></a></div><a id="slide-out-open" class="slide-out-open" href="#"><span></span></a><div id="fn-mc-ph"></div><div class="clear"></div><div id="main-nav" class="fixed-enabled"><nav id="navigation"><div class="main-menu novibrant"><div class="ul menu"><div class="menu-item menu-item-home"><a title="Aktuelle Nachrichten zu Aktien, Börse und Finanzen" href="https://www.finanznachrichten.de/">Startseite</a></div><div class="menu-item menu-item-has-children mega-menu mega-links "><a href="https://www.finanznachrichten.de/nachrichten/uebersicht.htm">Nachrichten</a><div class="mega-menu-block menu-sub-content"><div class="ul sub-menu-columns"><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Nachrichten</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/news.htm">Nachrichten auf FN</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/nachrichten-alle.htm">Alle News</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Rubriken</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/aktien-blickpunkt.htm">Aktien im Blickpunkt</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/ad-hoc-mitteilungen.htm">Ad hoc-Mitteilungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/bestbewertete-news.htm">Bestbewertete News</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/top-news.htm">Meistgelesene News</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/konjunktur-news.htm">Konjunktur- und Wirtschaftsnews</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/marktberichte.htm">Marktberichte</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/empfehlungen.htm">Empfehlungsübersicht</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/alle-empfehlungen.htm">Alle Aktienempfehlungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/chartanalysen.htm">Chartanalysen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/ipo-nachrichten.htm">IPO-News</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/termine.htm">Termine</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/dossiers/uebersicht.htm">Themen-Dossiers</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Indizes</a><div class="ul sub-menu-columns-item navi4col"><div class="menu-item full-width"><a title="Alle Indizes in der Übersicht" href="https://www.finanznachrichten.de/nachrichten-index/uebersicht.htm">Übersicht nach Indizes/Märkten</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/dax-30.htm"><span class="sprite flagge de"></span>DAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/sdax.htm"><span class="sprite flagge de"></span>SDAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/mdax.htm"><span class="sprite flagge de"></span>MDAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/tecdax.htm"><span class="sprite flagge de"></span>TecDAX</a></div><div class="menu-item clr fliess"><a href="https://www.finanznachrichten.de/nachrichten-index/dj-industrial.htm"><span class="sprite flagge us"></span>DJIA</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/s-p-500.htm"><span class="sprite flagge us"></span>S&amp;P 500</a></div><div class="menu-item half-width"><a href="https://www.finanznachrichten.de/nachrichten-index/nasdaq-100.htm"><span class="sprite flagge us"></span>NASDAQ 100</a></div><div class="menu-item clr fliess"><a href="https://www.finanznachrichten.de/nachrichten-index/euro-stoxx-50.htm"><span class="sprite flagge eu"></span>EURO STOXX 50</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/ftse-100.htm"><span class="sprite flagge gb"></span>FTSE-100</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/smi.htm"><span class="sprite flagge ch"></span>SMI</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/atx.htm"><span class="sprite flagge at"></span>ATX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/nikkei-225.htm"><span class="sprite flagge jp"></span>NIKKEI</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-index/hang-seng.htm"><span class="sprite flagge cn"></span>HANG SENG</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Branchen</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-branche/uebersicht.htm">Branchenübersicht</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Medien</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten-medien/uebersicht.htm">Medienübersicht</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/nachrichten/suche-medienarchiv.htm">Archiv</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item"><a href="https://www.finanznachrichten.de/suche/uebersicht.htm">Erweiterte Suche</a></div></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children mega-menu mega-links "><a href="https://www.finanznachrichten.de/aktienkurse/uebersicht.htm">Aktienkurse</a><div class="mega-menu-block menu-sub-content"><div class="ul sub-menu-columns"><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Aktienkurse</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse/realtime-kurse.htm">Realtime-Aktienkursliste (L&amp;S)</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/online-broker-vergleich.htm">Online-Broker-Vergleich</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">XETRA-Orderbuch</a><div class="ul sub-menu-columns-item navi2col"><div class="menu-item"><a href="https://aktienkurs-orderbuch.finanznachrichten.de/">Übersicht</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/xetra-orderbuch.htm">XETRA-Orderbuch?</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Indizes</a><div class="ul sub-menu-columns-item navi4col"><div class="menu-item full-width"><a href="https://www.finanznachrichten.de/aktienkurse/indizes.htm">Indexliste</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/dax-30.htm"><span class="sprite flagge de"></span>DAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/sdax.htm"><span class="sprite flagge de"></span>SDAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/mdax.htm"><span class="sprite flagge de"></span>MDAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/tecdax.htm"><span class="sprite flagge de"></span>TecDAX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/dj-industrial.htm"><span class="sprite flagge us"></span>DJIA</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/s-p-500.htm"><span class="sprite flagge us"></span>S&amp;P 500</a></div><div class="menu-item half-width"><a href="https://www.finanznachrichten.de/aktienkurse-index/nsadaq-100.htm"><span class="sprite flagge us"></span>NASDAQ 100</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/euro-stoxx-50.htm"><span class="sprite flagge eu"></span>EURO STOXX 50</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/ftse-100.htm"><span class="sprite flagge gb"></span>FTSE-100</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/smi.htm"><span class="sprite flagge ch"></span>SMI</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/atx.htm"><span class="sprite flagge at"></span>ATX</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/nikkei-225.htm"><span class="sprite flagge jp"></span>NIKKEI</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-index/hang-seng.htm"><span class="sprite flagge cn"></span>HANG SENG</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children"><a class="menue-heading" href="#">Branchen</a><div class="ul sub-menu-columns-item navi4col"><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/uebersicht.htm">Branchenübersicht</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/bau-infrastruktur.htm">Bau / Infrastrukur</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/bekleidung-textil.htm">Bekleidung / Textil</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/biotechnologie.htm">Biotechnologie</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/chemie.htm">Chemie</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/dienstleistungen.htm">Dienstleistungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/eisen-stahl.htm">Eisen / Stahl</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/elektrotechnologie.htm">Elektrotechnologie</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/erneuerbare-energien.htm">Erneuerbare Energien</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/fahrzeuge.htm">Fahrzeuge</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/finanzdienstleistungen.htm">Finanzdienstleistungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/freizeitprodukte.htm">Freizeitprodukte</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/gesundheitswesen.htm">Gesundheitswesen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/getraenke-tabak.htm">Getränke / Tabak</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/halbleiter.htm">Halbleiter</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/handel-e-commerce.htm">Handel / E-Commerce</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/hardware.htm">Hardware</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/holz-papier.htm">Holz / Papier</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/hotels-tourismus.htm">Hotels / Tourismus</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/immobilien.htm">Immobilien</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/industrie-mischkonzerne.htm">Industrie / Mischkonzerne</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/internet.htm">Internet</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/it-dienstleistungen.htm">IT-Dienstleistungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/konsumgueter.htm">Konsumgüter</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/kosmetik.htm">Kosmetik</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/kunststoffe-verpackungen.htm">Kunststoffe / Verpackungen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/logistik-transport.htm">Logistik / Transport</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/luftfahrt-ruestung.htm">Luftfahrt / Rüstung</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/maschinenbau.htm">Maschinenbau</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/medien.htm">Medien</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/nanotechnologie.htm">Nanotechnologie</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/nahrungsmittel-agrar.htm">Nahrungsmittel / Agrar</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/netzwerktechnik.htm">Netzwerktechnik</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/oel-gas.htm">Öl / Gas</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/pharma.htm">Pharma</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/rohstoffe.htm">Rohstoffe</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/software.htm">Software</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/sonstige-technologie.htm">Sonstige Technologie</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/telekom.htm">Telekommunikation</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/unterhaltung.htm">Unterhaltung</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/aktienkurse-branche/versorger.htm">Versorger</a></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item  "><a href="https://www.finanznachrichten.de/fonds/uebersicht.htm"><i class="icon fn-icon-banknote"></i>Fonds</a></div><div class="menu-item  "><a href="https://www.finanznachrichten.de/anleihen/uebersicht.htm"><i class="icon fn-icon-banknote"></i>Anleihen</a></div><div class="menu-item  "><a href="https://www.finanznachrichten.de/derivate/uebersicht.htm"><i class="icon fn-icon-banknote"></i>Derivate</a></div><div class="menu-item  "><a href="https://www.finanznachrichten.de/rohstoffe/uebersicht.htm"><i class="icon fn-icon-science-laboratory"></i>Rohstoffe</a></div><div class="menu-item menu-item-has-children mega-menu mega-links"><a href="https://www.finanznachrichten.de/devisen/uebersicht.htm"><i class="icon fn-icon-banknote"></i>Devisen</a><div class="mega-menu-block menu-sub-content"><div class="ul sub-menu-columns navi2col"><div class="menu-item full-width"><a title="Kryptowährungen" href="https://www.finanznachrichten.de/devisen/krypto-waehrungen.htm">Kryptowährungen</a></div></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div><div class="menu-item menu-item-has-children mega-menu mega-links"><a href="https://www.finanznachrichten.de/watchlist/nachrichten.htm"><i class="icon fn-icon-magnifying-glass"></i>Watchlist</a><div class="mega-menu-block menu-sub-content"><div class="ul sub-menu-columns navi2col"><div class="menu-item"><a href="https://www.finanznachrichten.de/watchlist/anlegen.htm">Watchlist anlegen</a></div><div class="menu-item"><a href="https://www.finanznachrichten.de/watchlist/information.htm">Was bringt eine Nachrichten-Watchlist?</a></div></div></div><i class="mobile-arrows icon fn-icon-down-open"></i></div></div></div></nav></div><div class="suche-mobile novibrant"><i class="icon fn-icon-magnifying-glass"></i></div><form action="https://www.finanznachrichten.de/suche/uebersicht.htm" method="get" id="fnkSucheForm" class="fnForm novibrant"><div id="sb-info"><span id="sb-datum">Donnerstag, 15.10.2020</span> Börsentäglich
        über 12.000 News von 649 internationalen
        Medien</div><div id="fnk-suche-cont"><div id="fnkSucheInputRow"><a id="fnkSucheErwL" class="fnChrome extendedSearch" href="https://www.finanznachrichten.de/suche/uebersicht.htm" data-baseurl="https://www.finanznachrichten.de/suche/uebersicht.htm?suche="><i class="icon fn-icon-search"></i>Erweiterte<br />Suche</a><i class="icon fn-icon-magnifying-glass"></i><input type="text" class="fnTextInput" id="fnk-suche-eingabe" accesskey="s" name="suche" required="required" /><input type="submit" id="fnk-suche-absenden" value="Suchen" /></div><div id="suchfeld-auto-vervollstaendigen" class="auto-vervollstaendigen"><div id="fnk-suche-werb"><span class="nadT extern href cursor_pointer" title="Die 1.400%-News: InnoCan Pharma: Phantastisches Zukunftsszenario f&#252;r Unternehmen und Investoren - strong buy!" data-nid="50959259" data-nad="1302">Phantastisches Zukunftsszenario f&#252;r Unternehmen und Investoren - strong buy!</span><div>Anzeige</div></div><table id="suchhilfeListe" class="suchHilfe"><tbody data-type="Indizes"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname"><br />Indizes</th><th class="sh-kurs"><br />Kurs</th><th class="sh-kurs"><br />%</th><th class="sh-news">News<br />24 h / 7 T</th><th class="sh-aufrufe">Aufrufe<br />7 Tage</th></tr><tr class="row_spacer"><td></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody data-type="Aktien"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname"><br />Aktien</th><th class="sh-kurs"><br />Kurs</th><th class="sh-kurs"><br />%</th><th class="sh-news">News<br />24 h / 7 T</th><th class="sh-aufrufe">Aufrufe<br />7 Tage</th></tr><tr class="row_spacer"><td></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody data-type="Orderbuch"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname colspan"><br />Xetra-Orderbuch</th><th class="sh-kurs"></th><th class="sh-kurs"></th><th class="sh-news"></th><th class="sh-aufrufe"></th></tr><tr class="row_spacer"><td></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname colspan"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody data-type="Devisen"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname"><br />Devisen</th><th class="sh-kurs"><br />Kurs</th><th class="sh-kurs"><br />%</th><th class="sh-news"></th><th class="sh-aufrufe"></th></tr><tr class="row_spacer"><td></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody data-type="Rohstoffe"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname"><br />Rohstoffe</th><th class="sh-kurs"><br />Kurs</th><th class="sh-kurs"><br />%</th><th class="sh-news"></th><th class="sh-aufrufe"></th></tr><tr class="row_spacer"><td></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody data-type="Dossier"><tr class="deco"><th class="sh-flagge"></th><th class="sh-prettyname"><br />Themen</th><th class="sh-kurs"><br />Kurs</th><th class="sh-kurs"><br />%</th><th class="sh-news"></th><th class="sh-aufrufe"></th></tr><tr class="row_spacer"><td colspan="6"></td></tr><tr class="hoverable template"><td class="sh-flagge"></td><td class="sh-prettyname"><span title=""></span></td><td class="sh-kurs"></td><td class="sh-kurs"></td><td class="sh-news"></td><td class="sh-aufrufe"></td></tr></tbody><tbody class="show"><tr class="row_spacer"><td colspan="6"></td></tr><tr><td class="sh-prettyname" colspan="6"><br /><a style="margin-left:20px;" class="fett extendedSearch" href="https://www.finanznachrichten.de/suche/uebersicht.htm" data-baseurl="https://www.finanznachrichten.de/suche/uebersicht.htm?suche=">Erweiterte Suche</a></td></tr></tbody></table></div></div></form><div id="sb-adhoc"><div id="sb-adhoc-bez"><strong><a class="fnChrome" href="https://www.finanznachrichten.de/nachrichten/ad-hoc-mitteilungen.htm">Ad hoc-Mitteilungen</a></strong>:</div><div id="sb-adhoc-mrq"><div id="sb-adhoc-mrq-c"><div id="sb-adhoc-mrq-s" class="marquee"><div id="sb-adhoc-mrq-sp" style="display: inline-block;"></div></div></div></div></div></div><div id="brotKrumen" class="bkContainer"><div id="bk" class="norm" itemprop="breadcrumb"><a rel="home" href="https://www.finanznachrichten.de/" title="FinanzNachrichten.de"><span class="sprite fn_vorstand icon_fn"></span></a><span><a href="https://www.finanznachrichten.de/aktienkurse/uebersicht.htm" title="Aktien"><span>Aktien</span></a><span class="brot">&raquo;</span><span><a href="https://www.finanznachrichten.de/nachrichten/uebersicht.htm" title="Nachrichten"><span>Nachrichten</span></a><span class="brot">&raquo;</span><span><a href="https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm" title="GENERIC GOLD AKTIE"><span>GENERIC GOLD AKTIE</span></a><span class="brot">&raquo;</span><span><a href="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm" title="Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million"><span>Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million</span></a></span></span></span></span></div><div id="ls"><button id="push-notifications-btn">Push Mitteilungen</button><span id="fn-als-startseite"><a href="https://www.finanznachrichten.de/service/fnalsstartseite.htm" title="Keine Nachrichten mehr verpassen: FinanzNachrichten.de zur Startseite machen">FN als Startseite</a></span><span id="RealtimeStatusSchalter" title="Hier klicken, um Realtime-Push-Kurse ein- oder auszuschalten"></span></div><div id="push-notifications"></div></div><div class="h-w-a"><div class="sprite h-w-b"></div><div id='dban1'></div><div id='sban1'></div><div id='yoc_top_HB'></div></div><div id="seitenbereiche"
     data-apihost="m.finanznachrichten.de"
     data-wsokw="GENERIC GOLD CORP,A2JAE9,CA37148M1068,1WD,GENERIC GOLD"
     data-wsopid="artikel_index_50275112"
     data-wsores="News"
     data-wsorub="News"
     data-wsourl="www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
     data-wsolay="Detailseite"
     data-wsotag="nachrichten"
     data-netpointfooter="False"><div class="seitenelement Aktien_Aktionen posoben main aktionen" id="W480" data-hasrealtime="False"><div class="widget-topCont clearfix bgDarkgray" itemscope="itemscope" itemtype="http://schema.org/Product" itemref="productid__47859 aktienbewertung_47859 ChartW480_2"><div class="c66 blockAt620 "><span class='aktie-zu-watchlist fn-icon-eye cursor_pointer' data-isin='CA37148M1068' title='GENERIC GOLD CORP zur Watchlist hinzufügen'></span><h2 class="widget-ueberschrift"><a href='https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm' itemprop='url'><span itemprop='name'>GENERIC GOLD CORP</span></a></h2><span id="productid__47859" class="wkn-isin"><span id="produkt-wkn" data-wkn="A2JAE9">WKN: A2JAE9<meta itemprop="productID" content="wkn:A2JAE9" /></span>&nbsp;<span id="produkt-isin" data-isin="CA37148M1068">ISIN: CA37148M1068</span>&nbsp;<span id="produkt-ticker" data-ticker="1WD">Ticker-Symbol: 1WD</span>&nbsp;</span></div><div id="partner-knockout"></div><span class="refresh-title-with-signalr" data-item="A2JAE9.FFM" data-ticker="1WD" data-previousclose="0.2820"></span><div class="c33 blockAt620 bgWhite"><div class="flex"><div><b itemprop="seller">Frankfurt</b></div><div class="border--left-right text--center"><span class="signalr" data-item="A2JAE9.FFM" data-field="Date_DateComponent" data-noflash="true">15.10.20</span></div><div class="text--right"><span class="signalr" data-item="A2JAE9.FFM" data-field="Date_SmallTimeComponent" data-noflash="true">09:29</span>&nbsp;Uhr<meta itemprop="priceValidUntil" content="2020-10-15T09:29:28.0000000" /></div></div><div class="flex flex--valign-center" id="additionalOfferInformation_47859"><div class="font--size-large flex--66" itemprop="offers" itemref="additionalOfferInformation_47859" itemscope="itemscope" itemtype="http://schema.org/Offer"><div data-item="A2JAE9.FFM" data-field="false" class="signalr sprite kgPfeil gleich"></div><span class="signalr grau" data-item="A2JAE9.FFM" data-field="false" data-noflash="true"><span class="signalr" data-item="A2JAE9.FFM" data-field="Rate">0,282<meta itemprop="price" content="0.28" /></span>&nbsp;Euro<meta itemprop="priceCurrency" content="EUR" /></span></div><div class="text--right"><span class="signalr text--no-wrap grau" data-item="A2JAE9.FFM" data-field="AbsDiff">0,000</span><br /><span class="signalr text--no-wrap grau" data-item="A2JAE9.FFM" data-field="RelDiff">0,00 %</span></div></div></div></div><ul class="tab-nav tab-nav--top-position owl-carousel"><li class="current" title="Nachrichten"><a href="https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm">Nachrichten</a></li><li title="Analysen"><a href="https://www.finanznachrichten.de/aktien-analysen/generic-gold-corp.htm">Analysen</a></li><li title="Kurse"><a href="https://www.finanznachrichten.de/aktienkurse-boersen/generic-gold-corp.htm">Kurse</a></li><li title="Chart"><a href="https://www.finanznachrichten.de/chart-tool/aktie/generic-gold-corp.htm">Chart</a></li></ul><div class="contentArea charts clearfix"><div class="data" id="cookieinfos" data-wkn="47859"></div><div class="a"><div class="u4">Branche</div><a href="https://www.finanznachrichten.de/nachrichten-branche/software.htm" title="Software" >Software</a><div class="u4">Aktienmarkt</div><a href="https://www.finanznachrichten.de/nachrichten-index/sonstige.htm" title="Sonstige" >Sonstige</a></div><div class="b"><div class="u4">1-Jahres-Chart</div><a href='https://www.finanznachrichten.de/chart-tool/aktie/generic-gold-corp.htm' id='ToolChartW480_2'> <img src='https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-1jahrklein-frankfurt.png' id='ChartW480_2' itemprop='image' width='180' height='100' alt='GENERIC GOLD CORP Chart 1 Jahr' title='GENERIC GOLD CORP Chart 1 Jahr' /></a></div><div class="c"><div class="u4">5-Tage-Chart</div><a href='https://www.finanznachrichten.de/chart-tool/aktie/generic-gold-corp.htm' id='ToolChart1_W480_1'> <img src='https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-intraklein-frankfurt.png' id='Chart1_W480_1' width='180' height='100' alt='GENERIC GOLD CORP 5-Tage-Chart' title='GENERIC GOLD CORP 5-Tage-Chart' /></a></div><div class="d"><div class="smartbroker--widget-stock-overview"><span class="on-click--partner sprite smartbroker-logo" data-val="smartbroker-stockoverview" title="Der Online Broker von Deutschlands größter Finanzcommunity"></span><span class="on-click--partner sprite smartbroker1" data-val="smartbroker-stockoverview" title="Jetzt für 0€ handeln!"></span></div></div></div></div><div class="sbSpalteL"><div class="seitenelement Nachrichten_Artikel poslinks" id="W459" data-hasrealtime="False"><div id="artikel_data" class="data" data-aktienzurnachricht="[[&quot;GENERIC GOLD CORP&quot;,&quot;https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm&quot;]]"></div><div itemscope="itemscope" itemtype="http://schema.org/Article"><meta itemprop="url" content="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm" /><meta itemprop="author" content="Newsfile" /><meta itemprop="datePublished" content="2020-07-27T14:22" /><meta itemprop="interactionCount" content="UserPageVisits:305" /><meta itemprop="keywords" content="Generic Gold Announces Upsizing Fully Subscribed Private Placement Million" /><meta itemprop="publisher" content="https://www.finanznachrichten.de" /><meta itemprop="headline" content="Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million" /><div id="article--rating-data" class="data" data-nachrichtid="50275112" data-cssclass="sterne45" data-feedurl="https://www.finanznachrichten.de/nachrichten-medien/newsfile.htm"></div><div id="article--header"><a href="https://www.finanznachrichten.de/nachrichten-medien/newsfile.htm"><strong>Newsfile</strong></a><div class="article--date-time">27.07.2020 | 14:22</div><div class="article--reader-count">305 Leser</div><div id="article--rating">Artikel bewerten:<div itemprop="aggregateRating" itemscope itemtype="http://schema.org/AggregateRating"><meta itemprop="itemReviewed" content="Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million" /><meta itemprop="worstRating" content="1" /><meta itemprop="bestRating" content="5" /><meta itemprop="ratingValue" content="4,5" /><div class="rateit" id="article--rating-stars" title="2 Leserbewertungen (4,5): unbedingt lesen"
                            data-rateit-value="4.5" data-rateit-resetable="false"
                            data-rateit-readonly="true" data-rateit-starwidth="12" data-rateit-starheight="12"></div><div id="article--rating-star-title"></div><div id="article--rating-title">(<span itemprop="ratingCount">2</span>)</div></div></div></div><h1 class="article--headline"><a href='https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm'>Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million</a></h1><div class="inhalte-text"><div itemprop="articleBody" id="artikelText" class="f_newsfile"><div id="artikelTextPuffer"><p>Toronto, Ontario--(Newsfile Corp. - July 27, 2020) - Generic Gold Corp. (CSE: GGC) (FSE: 1WD) ("<b>Generic Gold</b>" or the "<b>Company</b>") is pleased to announce, further to its press release of July 16, 2020, the upsize of its fully subscribed "best efforts" private placement offering, led by StephenAvenue Securities Inc. (the "<b>Agent</b>") as sole agent and sole bookrunner (the "<b>Offering</b>"), for aggregate gross proceeds of up to $7,000,000, through the issuance of units (each, a "<b>Unit</b>") at a price of $0.35 per Unit and flow-through units (each, a "<b>FT Unit</b>") at a price of $0.40 per FT Unit (together, the Units and the FT Units, the "<b>Offered Securities</b>").</p><p>The net proceeds from the sale of the Units will be used for general working capital and exploration purposes. The gross proceeds from the sale of the FT Units will be used by the Company to incur eligible "Canadian exploration expenses" that will qualify as "flow-through mining expenditures" (as such terms are defined in the <i>Income Tax Act</i> (Canada)) (the "<b>Qualifying Expenditures</b>") related to the Company's projects in Canada. All Qualifying Expenditures will be renounced in favour of the subscribers of the FT Units effective December 31, 2020. It is anticipated that most of the funds derived from the sale of the FT Units will be used to explore the Company's recently acquired Belvais project which is contiguous to Amex Exploration Inc.  (refer to the Company's press release of July 7, 2020).</p><p>The Offering is expected to close on or about August 6, 2020 (the "<b>Closing Date</b>"), or such other date as agreed between the Company and the Agent. The completion of the Offering is subject to certain closing conditions including, but not limited to, the receipt of all necessary regulatory and other approvals including the approval of the Canadian Securities Exchange. All Offered Securities will be subject to a statutory hold period of four months and one day from the Closing Date.</p><p>The Offered Securities have not been and will not be registered under the U.S. Securities Act of 1933, as amended, and may not be offered or sold in the United States absent registration or an applicable exemption from the registration requirements. This press release shall not constitute an offer to sell or the solicitation of an offer to buy nor shall there be any sale of the Offered Securities in any State in which such offer, solicitation or sale would be unlawful.<div id="nativendo-nachrichten-inarticle"></div></p><p><b>About Generic Gold</b></p><p>Generic Gold is a Canadian mineral exploration company focused on gold projects in the Abitibi Greenstone Belt in Quebec, Canada and Tintina Gold Belt in the Yukon Territory of Canada. The Company's Quebec exploration portfolio consists of three properties covering 8,148 hectares proximal to the town of Normétal and Amex Exploration's Perron project. The Company's Yukon exploration portfolio consists of several projects with a total land position of greater than 35,000 hectares, all of which are 100% owned by Generic Gold. Several of these projects are in close proximity to significant gold projects, including Goldcorp's Coffee project, Victoria Gold's Eagle Gold project, White Gold's Golden Saddle project, and Western Copper & Gold's Casino project. For information on the Company's property portfolio, visit the Company's website at genericgold.ca.</p><p><b>For further information contact: </b></p><p>Generic Gold Corp. <br />Richard Patricio, President and CEO <br />Tel: 416-456-6529 <br />rpatricio@genericgold.ca</p><p>StephenAvenue Securities Inc.<br />Daniel Cappuccitti<br />Tel: 416-479-4478<br />ecm@stephenavenue.com</p><p>NEITHER THE CANADIAN SECURITIES EXCHANGE NOR THEIR REGULATION SERVICES PROVIDERS ACCEPT RESPONSIBILITY FOR THE ADEQUACY OR ACCURACY OF THIS RELEASE.</p><p><i>Certain statements in this press release are "forward-looking" statements within the meaning of Canadian securities legislation. All statements, other than statements of historical fact, included herein are forward-looking information.  Forward-looking statements are necessarily based upon the current belief, opinions and expectations of management that, while considered reasonable by the Company, are inherently subject to business, economic, competitive, political and social uncertainties and other contingencies. Many factors could cause the Company's actual results to differ materially from those expressed or implied in the forward-looking statements. Accordingly, readers should not place undue reliance on forward-looking statements and forward-looking information. The Company does not undertake to update any forward-looking statements or forward-looking information that are incorporated by reference herein, except in accordance with applicable securities laws. Investors are cautioned not to put undue reliance on forward-looking statements due to the inherent uncertainty therein. We seek safe harbour.</i></p><p id="corporateNewsLogoContainer"><img src="https://orders.newsfilecorp.com/files/3923/60535_1df1c2fe4451d8fe_logo.jpg" id="corporateNewsLogo" alt="Corporate Logo" /></p><p id="corporateLinkBack">To view the source version of this press release, please visit https://www.newsfilecorp.com/release/60535</p><img width="1" height="1" style="width: 1px;  height: 1px;border: 0px solid;" src="https://www.newsfilecorp.com/newsinfo/60535/207" /></div><div class="smartbroker--text"><span title="Kostenloser Wertpapierhandel auf Smartbroker.de" class="nadT extern href cursor_pointer" data-nid="0" data-nad="652">GENERIC GOLD-Aktie komplett kostenlos handeln - auf Smartbroker.de</span></div></div><div class="article--copyright">&copy;&nbsp;2020&nbsp;<span onclick="FN.artikelKomplettID('0', 50275112)" class="href cursor_pointer">Newsfile Corp.</span></div><div class="article--footer"><a href="http://www.facebook.com/sharer.php?u=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
               title="Auf Facebook Teilen" target="_blank"><span title="Auf Facebook Teilen" class="fn-icon-fb"></span></a><a href="https://twitter.com/intent/tweet?source=webclient&url=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm%2F&text=Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million"
               title="Tweeten" target="_blank"><span title="Tweeten" class="fn-icon-twitter"></span></a><a href="https://www.xing.com/app/user?op=share;url=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
               title="Auf Xing Empfehlen" target="_blank"><span title="Auf Xing Empfehlen" class="fn-icon-xing"></span></a><a href="https://www.linkedin.com/shareArticle?mini=true&url=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm&title=deintitel"
               title="Auf LinkedIn Teilen" target="_blank"><span title="Auf LinkedIn Teilen" class="fn-icon-in"></span></a><a href="whatsapp://send?v=2&text=Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
               title="In WhatsApp Teilen" class="only-on-mobile"><span title="WhatsApp" class="fn-icon-whatsapp"></span></a><a href="https://share.flipboard.com/bookmarklet/popout?v=2&title=Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million&url=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
               title="FlipBoard" target="_blank"><span title="FlipBoard" class="fn-icon-flipboard"></span></a><a href="https://getpocket.com/edit?url=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
               title="GetPocket" target="_blank"><span title="GetPocket" class="fn-icon-getpocket"></span></a><span data-nid="50275112" class="fn-icon-print-2 on-click--print-article" title="Druckansicht"></span><span title="Schrift größer" class="fn-icon-font-incr on-click--incr-font-size"></span><span title="Schrift kleiner" class="fn-icon-font-decr on-click--decr-font-size"></span></div><div id="bookmarks_data" class="data"
             data-vibrant="True"
             data-nachrichtid="50275112"
             data-artikeladresse="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
             data-artikeladresseurlencoded="https%3a%2f%2fwww.finanznachrichten.de%2fnachrichten-2020-07%2f50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"></div></div></div></div><div id="dcb1"></div><div class="seitenelement Nachrichten_ZuAktie poslinks" id="W567_47859" data-hasrealtime="False"><div class="bedienfeld" id="W567_47859_progress"><span class="fn-icon-refresh cursor_pointer" id="W567_47859_pfad" data-xui-path="/w/567/47859?tab=0" title="Widget aktualisieren" data-idstring="W567_47859"></span></div><div class="widget-topCont"><div class="widget-ueberschrift"><a href='https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm'>Nachrichten zu GENERIC GOLD CORP</a></div></div><table class="rel-ct info einZ ft"><thead><tr><th class="Zeit" data-hide="" data-name="Zeit" >Zeit</th><th class="NewsTitel links" data-hide="" data-name="News">Aktuelle Nachrichten</th><th data-hide="" data-name="" data-toggle=&quot;True&quot;></th><th data-hide="all" data-name="" ></th></tr></thead><tbody class="table-hoverable table-alternating-rows"><tr><td class="zentriert"><span title='15:40'>29.09.</span></td><td class="zl"><span title="Generic Gold expedites acquisition of Belvais project in Quebec (MINING.com)" class="nT extern href cursor_pointer" data-nid="50826391">Generic Gold expedites acquisition of Belvais project in Quebec</span></td><td class=""></td><td class=""></td></tr><tr><td class="zentriert"><span title='13:58'>29.09.</span></td><td class="zl"><span title="Generic Gold Corp: Generic Gold acquires claim block at Belvais (Stockwatch)" class="nT extern href cursor_pointer" data-nid="50824912">Generic Gold Corp: Generic Gold acquires claim block at Belvais</span></td><td class=""></td><td class=""></td></tr><tr><td class="zentriert"><span title='13:21'>29.09.</span></td><td class="zl"><a href="https://www.finanznachrichten.de/nachrichten-2020-09/50824408-generic-gold-expands-belvais-land-position-in-normetal-region-quebec-296.htm" data-nid="50824408" title="Generic Gold Expands Belvais Land Position in Normetal Region, Quebec (Newsfile)" class="nT aufFn">Generic Gold Expands Belvais Land Position in Normetal Region, Quebec</a></td><td class=""></td><td class="">Toronto, Ontario--(Newsfile Corp. - September 29, 2020) - Generic Gold Corp. (CSE: GGC) (FSE: 1WD) ("Generic Gold" or the "Company") announces the acquisition (the "Transaction") of a large block of...<br /><span class="cursor_pointer href news_vorschau" data-nid="50824408">&#x25ba; Artikel lesen</span></td></tr><tr><td class="zentriert"><span title='13:23'>28.09.</span></td><td class="zl"><span title="Generic Gold Corp: Generic Gold speeds up acquisition of Belvais (Stockwatch)" class="nT extern href cursor_pointer" data-nid="50812879">Generic Gold Corp: Generic Gold speeds up acquisition of Belvais</span></td><td class=""></td><td class=""></td></tr><tr><td class="zentriert"><span title='13:22'>28.09.</span></td><td class="zl"><a href="https://www.finanznachrichten.de/nachrichten-2020-09/50812829-generic-gold-expedites-property-acquisition-in-abitibi-region-of-quebec-296.htm" data-nid="50812829" title="Generic Gold Expedites Property Acquisition in Abitibi Region of Quebec (Newsfile)" class="nT aufFn">Generic Gold Expedites Property Acquisition in Abitibi Region of Quebec</a></td><td class=""></td><td class="">Toronto, Ontario--(Newsfile Corp. - September 28, 2020) -  Generic Gold Corp. (CSE: GGC) (FSE: 1WD) ("Generic Gold" or the "Company") announces that further to its press release of July 7, 2020, the...<br /><span class="cursor_pointer href news_vorschau" data-nid="50812829">&#x25ba; Artikel lesen</span></td></tr></tbody></table></div></div><div class="sbSpalteR"><div id="dmr1"></div><div id="yoc_intext_middle_1_HB"></div><div id="yoc_general_middle_1_HB"></div><div id="sban2"></div><div class="seitenelement Aktien_AusArtikel posrechts" id="W69" data-hasrealtime="False"><div class="bedienfeld" id="W69_progress"><span class="fn-icon-refresh cursor_pointer" id="W69_pfad" data-xui-path="/w/69/50275112?tab=0" title="Widget aktualisieren" data-idstring="W69"></span></div><div class="widget-topCont"><div class="widget-ueberschrift">Firmen im Artikel</div></div><div class="inhalte-text zentriert"><div class="u4" id="W69_article50275112_Chart_Titel">5-Tage-Chart<br />GENERIC GOLD</div><a href='https://www.finanznachrichten.de/chart-tool/aktie/generic-gold-corp.htm' id='ToolW69_article50275112_Chart'> <img src='https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-intraday-frankfurt.png' id='W69_article50275112_Chart' width='230' height='200' alt='GENERIC GOLD CORP Chart 1 Jahr' title='GENERIC GOLD CORP Chart 1 Jahr' /></a></div><table class="info einZ table-hoverable table-alternating-rows"><thead><tr><th><span data-xui-path="/w/69/50275112?tab=0&amp;sort=Firma_ab" class="fnxui spfc auf aktiv" title="Firma (aufsteigend)">Unternehmen / Aktien<span class="sprite sortierpfeil"></span></span></th><th class="w62"><span data-xui-path="/w/69/50275112?tab=0&amp;sort=Kurs_ab" class="fnxui spfc ab" title="Kurs">Kurs<span class="sprite sortierpfeil"></span></span></th><th class="w62"><span data-xui-path="/w/69/50275112?tab=0&amp;sort=diffrel_ab" class="fnxui spfc ab" title="diffrel">%<span class="sprite sortierpfeil"></span></span></th></tr></thead><tbody><tr><td class="cwHoverChart"
                            data-chart="https://www.finanznachrichten.de/chart-generic-gold-corp-aktie-intraday-frankfurt.png"
                            data-charttool="https://www.finanznachrichten.de/chart-tool/aktie/generic-gold-corp.htm"
                            data-hoverchart="W69_article50275112_Chart"
                            data-hovertitel="W69_article50275112_Chart_Titel"
                            data-charttitel="5-Tage-Chart
GENERIC GOLD"
                            title="GENERIC GOLD CORP"><a href="https://www.finanznachrichten.de/nachrichten-aktien/generic-gold-corp.htm" title="GENERIC GOLD CORP" class="fnAktie" >GENERIC GOLD CORP</a></td><td class="rechts">0,282</td><td class="rechts"><span class="aktie-gleich">0,00&nbsp;%</span></td></tr></tbody></table></div><div id="WAd_NewsKnockouts" class="lazy-widget" data-path="/w/ad_newsknockouts/50275112?tab=0"></div><div id="W70" class="lazy-widget" data-path="/w/70/50275112?tab=4"></div><div id="W78" class="lazy-widget" data-path="/w/78/50275112?tab=0"></div><div id="W73" class="lazy-widget" data-path="/w/73/50275112?tab=0"></div></div><div class="sbSpalteL"></div><div class="sbSpalteR"></div></div><div id='sist1'></div><div id='yoc_outofpage_HB'></div><div id="fuss novibrant"><div id="fusszeile-links" class="seitenelement main">Sie erhalten auf FinanzNachrichten.de kostenlose Realtime-Aktienkurse von <span class="sprite ls cursor_pointer on-click--partner" title="Lang &amp; Schwarz" data-val="lang-und-schwarz"></span>und <span class="sprite tg cursor_pointer on-click--partner" title="Tradegate" data-val="tradegate"></span></div><div class="clr"></div><div id="dban2"></div><div id="cban1"></div><div id="yoc_bottom_HB"></div><div id="footer-sozial"><div id="footer_version">FNRD-2.619.0</div><div id="seiten-bewertung"><div>Wie bewerten Sie die aktuell angezeigte Seite?</div><span class="vorheriger">sehr gut</span><span class="buttons" data-val='/service/seitenbewertung.htm'><span class="on-click--rate-site">1</span><span class="on-click--rate-site">2</span><span class="on-click--rate-site">3</span><span class="on-click--rate-site">4</span><span class="on-click--rate-site">5</span><span class="on-click--rate-site">6</span></span><span class="naechster">schlecht</span><span class="buttons btnProblem"><span style="margin-left: 15px" class="on-click--show-fancy-box" data-val="/service/problembericht.htm" title="Problem melden">Problem melden</span></span></div><div id="footer_social"><div class='facebookbuttons social-media-wrapper'><a class='sprite icon icon-flike facebookfanwerden' title='Jetzt Fan werden!' href='https://www.facebook.com/finanznachrichten'></a><a class='fb-share sprite empfehlenimage' target='_blank' href='https://www.facebook.com/sharer/sharer.php?u=https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm'></a></div><div class='twitterbutton social-media-wrapper'><div class='sprite twitterimage cursor_pointer' data-url='https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm' data-text='Generic Gold Announces Upsizing of Fully Subscribed Private Placement up to $7 Million'></div></div></div><div class="clear"></div></div><div class="sb-disclaimer"><b>Werbehinweise: </b>Die Billigung des Basisprospekts durch die BaFin ist nicht als ihre Befürwortung der angebotenen Wertpapiere zu verstehen. 
    Wir empfehlen Interessenten und potenziellen Anlegern den Basisprospekt und die Endgültigen Bedingungen zu lesen, 
    bevor sie eine Anlageentscheidung treffen, um sich möglichst umfassend zu informieren, insbesondere über die potenziellen 
    Risiken und Chancen des Wertpapiers. Sie sind im Begriff, ein Produkt zu erwerben, das nicht einfach ist und schwer zu 
    verstehen sein kann.
</div><div id="fusszeile" class="footer-navi"><div class="footer-scroll-top" title="Nach oben"><i class="icon fn-icon-scroll-top"></i></div><div class="footer-logo"><div></div><a href="/"><img src="https://fns1.de/img/logo_slogan.svg" alt="FinanzNachrichten.de" /></a></div><div class="footer-top"><div class="lnks"><a href="https://www.finanznachrichten.de/aktienkurse-index/dax-30.htm" title="DAX">DAX</a><a href="https://aktienkurs-orderbuch.finanznachrichten.de/" title="Xetra-Orderbuch">Xetra-Orderbuch</a><a href="https://www.finanznachrichten.de/nachrichten/ad-hoc-mitteilungen.htm" title="Ad hoc-Mitteilungen">Ad hoc-Mitteilungen</a><a href="https://www.finanznachrichten.de/nachrichten-index/uebersicht.htm" title="Nachrichten Börsen">Nachrichten Börsen</a><a href="https://www.finanznachrichten.de/nachrichten/empfehlungen.htm" title="Aktien-Empfehlungen">Aktien-Empfehlungen</a><a href="https://www.finanznachrichten.de/nachrichten-branche/uebersicht.htm" title="Branchen">Branchen</a><a href="https://www.finanznachrichten.de/nachrichten-medien/uebersicht.htm" title="Medien">Medien</a><a href="https://www.finanznachrichten.de/nachrichten/suche-medienarchiv.htm" title="Nachrichten-Archiv">Nachrichten-Archiv</a><div class="rss"><a href="https://www.finanznachrichten.de/service/rss.htm">RSS-News von FinanzNachrichten.de</a></div></div></div><div class="footer-bottom"><a href="https://www.finanznachrichten.de/service/presse.htm" rel="nofollow" title="Presse" class="fst">Presse</a><a href="https://www.finanznachrichten.de/service/impressum.htm" rel="nofollow" title="Impressum | AGB | Disclaimer | Datenschutz">Impressum | AGB | Disclaimer | Datenschutz</a><a href="https://www.finanznachrichten.de/service/mediadaten.htm" rel="nofollow" title="Mediadaten" class="lst"><span class="fett">Mediadaten</span></a></div></div></div><div id="smartbroker--head"><span class="on-click--partner sprite smartbroker-logo" data-val="smartbroker-head" title="Der Online Broker von Deutschlands größter Finanzcommunity"></span><span class="on-click--partner sprite smartbroker1" data-val="smartbroker-head" title="Jetzt für 0€ handeln!"></span></div></div><div id="url-css" class="data" data-all="https://fns1.de/css/fn216.css"></div><div id="model_data" class="data" data-realtime="False" data-tickernews="True"></div><div id="signalr_data" class="data" data-url="https://rt.finanznachrichten.de/signalr"></div><div id="ivw_data" class="data" data-zaehlcode="news-01"></div><div id="suchhilfe_data" class="data" data-url="https://www.finanznachrichten.de"></div><noscript><img height="1" width="1" style="display:none" src="https://www.facebook.com/tr?id=798877413578193&ev=PageView&noscript=1" /></noscript><span id="ads_data" class="data"
      data-targeting-url="https://www.finanznachrichten.de/nachrichten-2020-07/50275112-generic-gold-announces-upsizing-of-fully-subscribed-private-placement-up-to-dollar-7-million-296.htm"
      data-targeting-keywords="[&quot;Generic&quot;,&quot;Gold&quot;,&quot;Announces&quot;,&quot;Upsizing&quot;,&quot;of&quot;,&quot;Fully&quot;,&quot;Subscribed&quot;,&quot;Private&quot;,&quot;Placement&quot;,&quot;up&quot;,&quot;to&quot;,&quot;$7&quot;,&quot;Million&quot;]"
      data-targeting-bereich="artikel_index"
      data-topbanner="True"
      data-articlebanner="True"
      data-leftsidebanner="False"></span><div id='div-gpt-ad-1479204326049-4'></div><a id="ads_wp" href="#" target="_blank"></a><div id="dsky1"></div><div id="dsky"></div></div><div class="modal fade" id="dialogbox" tabindex="-1" role="dialog" aria-hidden="true"><div class="modal-dialog"><div class="modal-content" id="dialogboxcontent"></div></div></div><script type="text/javascript">(function(){var t=document.getElementsByTagName("head")[0]||document.documentElement,n=document.createElement("script"),i;n.src="https://fns1.de/js/foot200.js";i=!1;n.onload=n.onreadystatechange=function(){i||this.readyState&&this.readyState!=="loaded"&&this.readyState!=="complete"||(i=!0,n.onload=n.onreadystatechange=null,t&&n.parentNode&&t.removeChild(n))};t.insertBefore(n,t.firstChild)})()</script><script language="javascript" src="/scripts/8c584ae5ed8f9fd2a1bd61af57826caa3438ae3c.js"></script><script src="https://www.gstatic.com/firebasejs/6.3.3/firebase-app.js"></script><script src="https://www.gstatic.com/firebasejs/6.3.3/firebase-messaging.js"></script><script>firebase.initializeApp({
            apiKey: "AIzaSyAgKusdydsLP0kHKJjBV7bGyg18uJZnjBo",
            authDomain: "finanznachrichten-8126c.firebaseapp.com",
            databaseURL: "https://finanznachrichten-8126c.firebaseio.com",
            projectId: "finanznachrichten-8126c",
            storageBucket: "finanznachrichten-8126c.appspot.com",
            messagingSenderId: "303111472022",
            appId: "1:303111472022:web:6d8b3fe9658bf905"
        });</script><!-- Quantcast Tag --><script type="text/javascript">var _qevents = _qevents || [];

            (function () {
                var elem = document.createElement('script');
                elem.src = (document.location.protocol == "https:" ? "https://secure" : "http://edge") + ".quantserve.com/quant.js";
                elem.async = true;
                elem.type = "text/javascript";
                var scpt = document.getElementsByTagName('script')[0];
                scpt.parentNode.insertBefore(elem, scpt);
            })();

            _qevents.push({
                qacct: "p-XQPkmN_wVNAmT"
            });</script><noscript><div style="display:none;"><img src="//pixel.quantserve.com/pixel/p-XQPkmN_wVNAmT.gif" border="0" height="1" width="1" alt="Quantcast" /></div></noscript><!-- End Quantcast tag --></body></html>
`
