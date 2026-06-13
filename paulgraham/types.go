package paulgraham

import (
	"strings"
)

// Essay is the record for a Paul Graham essay listing.
type Essay struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// EssayContent holds the full content of a fetched essay.
type EssayContent struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body"`
}

// parseArticlesHTML parses the /articles.html page and returns all essay links.
// It looks for href="*.html" anchors that are local (no "://").
func parseArticlesHTML(html, baseURL string) []Essay {
	var essays []Essay
	rest := html
	for {
		// Find next <a href="
		idx := strings.Index(strings.ToLower(rest), "<a href=\"")
		if idx < 0 {
			break
		}
		rest = rest[idx+len("<a href=\""):]

		// Find closing quote for href value
		end := strings.IndexByte(rest, '"')
		if end < 0 {
			break
		}
		href := rest[:end]
		rest = rest[end+1:]

		// Skip external links and non-.html links
		if strings.Contains(href, "://") || !strings.HasSuffix(href, ".html") {
			continue
		}
		// Skip links that look like full paths (with directory separators)
		if strings.Contains(href, "/") {
			continue
		}

		slug := strings.TrimSuffix(href, ".html")
		if slug == "" {
			continue
		}

		// Find anchor text (between > and </a>)
		gtIdx := strings.IndexByte(rest, '>')
		if gtIdx < 0 {
			continue
		}
		rest = rest[gtIdx+1:]

		closeIdx := strings.Index(strings.ToLower(rest), "</a>")
		if closeIdx < 0 {
			continue
		}
		title := stripTags(rest[:closeIdx])
		rest = rest[closeIdx:]

		if title == "" {
			title = slug
		}

		essays = append(essays, Essay{
			Slug:  slug,
			Title: title,
			URL:   baseURL + "/" + href,
		})
	}
	return essays
}

// parseEssayHTML extracts the title and body text from an essay HTML page.
func parseEssayHTML(html, slug, url string) EssayContent {
	title := extractTitle(html, slug)
	body := extractBody(html)
	return EssayContent{
		Slug:  slug,
		Title: title,
		URL:   url,
		Body:  body,
	}
}

// extractTitle tries to find the essay title from <title> or <h3> tags.
func extractTitle(html, slug string) string {
	lower := strings.ToLower(html)

	// Try <title>...</title>
	start := strings.Index(lower, "<title>")
	if start >= 0 {
		rest := html[start+len("<title>"):]
		end := strings.Index(strings.ToLower(rest), "</title>")
		if end >= 0 {
			t := strings.TrimSpace(rest[:end])
			if t != "" {
				return t
			}
		}
	}

	// Try <h3>...</h3>
	start = strings.Index(lower, "<h3>")
	if start >= 0 {
		rest := html[start+len("<h3>"):]
		end := strings.Index(strings.ToLower(rest), "</h3>")
		if end >= 0 {
			t := strings.TrimSpace(stripTags(rest[:end]))
			if t != "" {
				return t
			}
		}
	}

	return slug
}

// extractBody extracts plain text from the essay body. PG essays use
// <font face=verdana> or <table> based layouts; we strip all tags.
func extractBody(html string) string {
	// Strip the <head>...</head> section
	lower := strings.ToLower(html)
	headEnd := strings.Index(lower, "</head>")
	if headEnd >= 0 {
		html = html[headEnd+len("</head>"):]
	}

	text := stripTags(html)

	// Collapse multiple blank lines into one
	lines := strings.Split(text, "\n")
	var out []string
	blank := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blank++
			if blank <= 1 {
				out = append(out, "")
			}
		} else {
			blank = 0
			out = append(out, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// stripTags removes HTML tags and decodes common entities.
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
			b.WriteRune('\n')
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	out := b.String()
	out = strings.ReplaceAll(out, "&amp;", "&")
	out = strings.ReplaceAll(out, "&lt;", "<")
	out = strings.ReplaceAll(out, "&gt;", ">")
	out = strings.ReplaceAll(out, "&quot;", `"`)
	out = strings.ReplaceAll(out, "&#39;", "'")
	out = strings.ReplaceAll(out, "&apos;", "'")
	out = strings.ReplaceAll(out, "&nbsp;", " ")
	out = strings.ReplaceAll(out, "&#160;", " ")
	return strings.TrimSpace(out)
}
