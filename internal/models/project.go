package models

type Project struct {
	BuildDir      string `json:"target_dir"`
	RootDir       string `json:"root_dir"`
	AppMainSrcDir string `json:"app_main_src_dir"`
	AppName       string `json:"app_name"`
}
