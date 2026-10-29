# Tech Stack


1. Golang
2. React pages for web (embeded in golang)
3. Bleve v2 (pure Go) for full-text search of md files. Fuzzy (edit distance 1 for words <= 5 chars, 2 for longer), fields searched: title, name, description, tags, content, section, project (boosts: title and name 3, description and tags 2, others 1), default limit 20. A schema version change rebuilds the index on server start.
