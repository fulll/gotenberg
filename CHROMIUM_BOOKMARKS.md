# Custom PDF outlines in Chromium

The Chromium URL, HTML, and Markdown PDF conversion routes accept the fork's
`bookmarks` multipart form field. Pass a pdfcpu bookmark tree as JSON and set
`generateDocumentOutline=true`:

```shell
curl --form 'files=@index.html' \
  --form 'generateDocumentOutline=true' \
  --form 'bookmarks={"bookmarks":[{"title":"Chapter","page":1,"kids":[{"title":"Section","page":2}]}]}' \
  http://localhost:3000/forms/chromium/convert/html --output output.pdf
```

Pages are one-based; the example expects at least two pages. Use `page` (not
`PageFrom`) and `kids` for nested entries. As in the original fork, omitted or
empty bookmark trees skip import, and custom import does not overwrite an
outline already emitted by Chromium. If it emits an outline for your source,
use Gotenberg's native PDF set-bookmarks route to replace it instead.

The `--pdfengines-import-bookmarks-engines` flag
(`PDFENGINES_IMPORT_BOOKMARKS_ENGINES`) selects the engines for this import,
independently of the upstream PDF read/write bookmark routes.
