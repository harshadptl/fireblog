package handlers

import (
	"encoding/xml"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	AtomLink      atomLink  `xml:"atom:link"`
	Items         []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	GUID        string   `xml:"guid"`
	Description string   `xml:"description"`
	Author      string   `xml:"dc:creator,omitempty"`
	PubDate     string   `xml:"pubDate"`
	Content     cdataStr `xml:"content:encoded"`
}

type cdataStr struct {
	Value string `xml:",cdata"`
}

// RSS serves the latest posts as an RSS 2.0 feed.
func (h *Handler) RSS(c fiber.Ctx) error {
	posts, err := h.Store.AllPublishedForFeed(c, 20)
	if err != nil {
		return err
	}
	base := h.Cfg.BaseURL
	ch := rssChannel{
		Title:       h.Cfg.AppName,
		Link:        base + "/",
		Description: "Latest posts from " + h.Cfg.AppName,
		Language:    "en",
		AtomLink:    atomLink{Href: base + "/feed.xml", Rel: "self", Type: "application/rss+xml"},
	}
	for i, p := range posts {
		if i == 0 {
			ch.LastBuildDate = p.PublishedAt.UTC().Format(time.RFC1123Z)
		}
		link := base + "/posts/" + url.PathEscape(p.Slug)
		item := rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			Description: p.Excerpt,
			PubDate:     p.PublishedAt.UTC().Format(time.RFC1123Z),
			Content:     cdataStr{strings.ReplaceAll(p.ContentHTML, "]]>", "]]&gt;")},
		}
		if p.Author != nil {
			item.Author = p.Author.Name
		}
		ch.Items = append(ch.Items, item)
	}

	out, err := xml.MarshalIndent(rssFeed{Version: "2.0", Atom: "http://www.w3.org/2005/Atom", Channel: ch}, "", "  ")
	if err != nil {
		return err
	}
	// Namespaces for dc: and content: are declared by string replacement to
	// keep the structs simple.
	doc := strings.Replace(string(out), `<rss version="2.0"`,
		`<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/"`, 1)
	c.Set(fiber.HeaderContentType, "application/rss+xml; charset=utf-8")
	c.Set(fiber.HeaderCacheControl, "public, max-age=600")
	return c.SendString(xml.Header + doc)
}

type urlset struct {
	XMLName xml.Name     `xml:"urlset"`
	NS      string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// Sitemap lists public URLs for search engines.
func (h *Handler) Sitemap(c fiber.Ctx) error {
	posts, err := h.Store.AllPublishedForFeed(c, 5000)
	if err != nil {
		return err
	}
	tags, err := h.Store.PopularTags(c, 0)
	if err != nil {
		return err
	}
	base := h.Cfg.BaseURL
	set := urlset{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	set.URLs = append(set.URLs, sitemapURL{Loc: base + "/"}, sitemapURL{Loc: base + "/tags"})
	authors := map[string]bool{}
	for _, p := range posts {
		set.URLs = append(set.URLs, sitemapURL{
			Loc:     base + "/posts/" + url.PathEscape(p.Slug),
			LastMod: p.UpdatedAt.UTC().Format("2006-01-02"),
		})
		if p.Author != nil && !authors[p.Author.Username] {
			authors[p.Author.Username] = true
			set.URLs = append(set.URLs, sitemapURL{Loc: base + "/u/" + url.PathEscape(p.Author.Username)})
		}
	}
	for _, t := range tags {
		set.URLs = append(set.URLs, sitemapURL{Loc: base + "/tags/" + url.PathEscape(t.Slug)})
	}
	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "application/xml; charset=utf-8")
	c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	return c.SendString(xml.Header + string(out))
}

// Robots serves robots.txt.
func (h *Handler) Robots(c fiber.Ctx) error {
	body := "User-agent: *\n" +
		"Disallow: /dashboard\nDisallow: /settings\nDisallow: /admin\nDisallow: /auth/\n" +
		"\nSitemap: " + h.Cfg.BaseURL + "/sitemap.xml\n"
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return c.SendString(body)
}
