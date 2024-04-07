package dirindex

import (
  "context"
  "errors"
  "embed"
  "html/template"
  "io/fs"
  "log"
  "os"
  "path/filepath"
  "strings"
  "github.com/emad-elsaid/xlog"
)

//go:embed templates
var templates embed.FS

func init() {
  xlog.RegisterTemplate(templates, "templates")
  xlog.RegisterPageSource(new(dirSource))
}

type dirPage struct {
  name string
  xlog.BasePage
}

type page struct {
  Name string
  Link string
}

func (p *dirPage) Content() xlog.Markdown {
  pages := []page{}

  // https://stackoverflow.com/a/71172704
  sepCount := strings.Count(p.FileName(), string(os.PathSeparator))
  filepath.WalkDir(p.FileName(), func(name string, d fs.DirEntry, err error) error {
    if err != nil {
      return err
    }
    pages = append(pages, page{
      name,
      strings.TrimSuffix(name, ".md"),  // TODO: what about html?
    })
    if d.IsDir() && strings.Count(name, string(os.PathSeparator)) > sepCount {
      // no need to recurse into subdirs
      return fs.SkipDir
    }
    return nil
  })
  return xlog.Markdown(xlog.Partial("dirindex", xlog.Locals{
    "pages": pages,
  }))
}

func (p *dirPage) FileName() string {
  return filepath.FromSlash(strings.TrimSuffix(p.name, "/"))
}

func (p *dirPage) Exists() bool {
  _, err := os.Stat(p.FileName())
  if err != nil {
    return false
  }
  return true
}

func (p *dirPage) Render() template.HTML {
  return template.HTML(p.Content())
}

type dirSource struct {
  xlog.PageSource
}

 func (p *dirSource) Page(name string) xlog.Page {
  fileInfo, err := os.Stat(filepath.FromSlash(strings.TrimSuffix(name, "/")))
  if (err != nil || ! fileInfo.IsDir()) {
    return nil
  }
  log.Printf("DirIndex: %s", name)
  return &dirPage{name: name}
}

// Needed when creating files
func (p *dirSource) Each(ctx context.Context, f func(xlog.Page)) {
  filepath.WalkDir(".", func(name string, d fs.DirEntry, err error) error {
    if err != nil {
      return err
    }
		select {

		case <-ctx.Done():
			return errors.New("context stopped")

		default:
      if ! d.IsDir() {
        return nil
      }

      // TODO: is this needed? If so, support other extensions
      _, err := os.Stat(filepath.FromSlash(name) + "index.md")
      if err != nil {
        // let other page sources handle directories with indexes
        return nil
      }

      f(&dirPage{
        name: name,
      })
		}

		return nil
	})
}
