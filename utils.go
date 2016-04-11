package main

import (
	"fmt"
)

type Xkcd struct {
	Title string `json:"title"`
	Image string `json:"img"`
}

func getFormat(contents []byte) string {
	if len(contents) < 4 {

		fmt.Println("Not enough bytes")
		return ""
	}

	if contents[0] == 0x89 && contents[1] == 0x50 &&
		contents[2] == 0x4E && contents[3] == 0x47 {
		return "png"
	}

	if contents[0] == 0xFF && contents[1] == 0xD8 {
		return "jpg"
	}

	if contents[0] == 0x47 && contents[1] == 0x49 &&
		contents[2] == 0x46 && contents[3] == 0x38 {
		return "gif"
	}

	if contents[0] == 0x42 && contents[1] == 0x4D {
		return "bmp"
	}

	return ""
}
