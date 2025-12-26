package context

type Context struct {
	ImageName   string
	DownloadDir string
}

func New(imageName, downloadDir string) *Context {
	return &Context{
		ImageName:   imageName,
		DownloadDir: downloadDir,
	}
}
