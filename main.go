package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/0gajun/esa-archiver/esa"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Please specify a target directory")
		os.Exit(1)
	}
	targetDir := os.Args[1]

	accessToken, ok := os.LookupEnv("ESA_TOKEN")
	if !ok {
		fmt.Printf("Missing ESA_TOKEN environment variable")
		os.Exit(1)
	}

	esaClient, err := esa.NewEsa(accessToken, "0gajun")
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	posts, err := esaClient.GetAllPosts(context.Background())
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	epa := newEsaPostArchiver(targetDir)
	for _, post := range posts {
		if err := epa.do(&post); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}
}

const attachmentRe = `\[([^\[\]]+) (\([^\(\)]+)\)\]\(([^\(\)]+)\)`

type esaPostArchiver struct {
	re         *regexp.Regexp
	archiveDir string

	esaAttachmentUrlPrefix []string
}

func newEsaPostArchiver(archiveDir string) *esaPostArchiver {
	return &esaPostArchiver{
		re:         regexp.MustCompile(attachmentRe),
		archiveDir: archiveDir,
		esaAttachmentUrlPrefix: []string{
			"https://img.esa.io/",
			"https://esa-storage-tokyo.s3-ap-northeast-1.amazonaws.com/",
		},
	}
}

type attachment struct {
	filename string
	url      string
}

func (e *esaPostArchiver) findAllAttachments(body string) []attachment {
	attachments := []attachment{}

	results := e.re.FindAllStringSubmatch(body, -1)
	for _, result := range results {
		filename := result[1]
		url := result[3]
		if e.isAttachment(url) {
			attachments = append(attachments, attachment{filename: filename, url: url})
		}
	}

	return attachments
}

func (e *esaPostArchiver) do(post *esa.Post) error {
	dirs := strings.Split(post.FullName, "/")
	dirs = dirs[0 : len(dirs)-1] // Remove post title

	targetDirPath := fmt.Sprintf("%s/%s", e.archiveDir, strings.Join(dirs, "/"))
	if err := os.MkdirAll(targetDirPath, os.ModeDir|0755); err != nil {
		return err
	}

	replacerArg := []string{}

	a := func(attachments []attachment) ([]string, error) {
		replacerArg := []string{}

		for _, at := range attachments {
			relativeFilePath := fmt.Sprintf("./attachments/%s", at.filename)
			absoluteFilePath := fmt.Sprintf("%s/%s", targetDirPath, relativeFilePath)
			os.MkdirAll(fmt.Sprintf("%s/attachments", targetDirPath), os.ModeDir|0755)
			out, err := os.Create(absoluteFilePath)
			if err != nil {
				return []string{}, err
			}
			defer out.Close()

			resp, err := http.Get(at.url)
			if err != nil {
				return []string{}, err
			}
			defer resp.Body.Close()

			if _, err := io.Copy(out, resp.Body); err != nil {
				return []string{}, err
			}

			replacerArg = append(replacerArg, at.url, relativeFilePath)
		}

		return replacerArg, nil
	}

	attachments := e.findAllAttachments(post.BodyMd)
	result, err := a(attachments)
	if err != nil {
		return err
	}
	replacerArg = append(replacerArg, result...)

	r := strings.NewReplacer(replacerArg...)
	replacedBody := r.Replace(post.BodyMd)

	out, err := os.Create(fmt.Sprintf("%s/%s.md", targetDirPath, post.Name))
	if err != nil {
		return nil
	}
	defer out.Close()

	io.WriteString(out, replacedBody)

	if len(post.Comments) > 0 {
		io.WriteString(out, "\n# -------Comments--------\n")

		for _, comment := range post.Comments {
			attachments := e.findAllAttachments(comment.BodyMd)
			result, err := a(attachments)
			if err != nil {
				return err
			}
			replacerArg = append(replacerArg, result...)
			r := strings.NewReplacer(replacerArg...)
			replacedBody := r.Replace(comment.BodyMd)

			io.WriteString(out, fmt.Sprintf("### %s\n", comment.CreatedAt.String()))
			io.WriteString(out, replacedBody)
			io.WriteString(out, "\n")
		}
	}

	return nil
}

func (e *esaPostArchiver) isAttachment(url string) bool {
	for _, prefix := range e.esaAttachmentUrlPrefix {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}
