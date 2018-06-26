package services

import (
	"gopkg.in/gographics/imagick.v2/imagick"
)

type Compression struct {
	Wand *imagick.MagickWand
}

func NewMagickWand() Compression {

	return Compression{
		Wand: imagick.NewMagickWand(),
	}
}

func (comp *Compression) Close() {
	comp.Wand.Destroy()
}

func (comp *Compression) ResizeImage(img []byte) []byte {
	comp.Wand.ReadImageBlob(img)
	comp.Wand.SetCompressionQuality(85)
	comp.Wand.ResizeImage(
		comp.Wand.GetImageWidth()/2,
		comp.Wand.GetImageHeight()/2,
		imagick.FILTER_UNDEFINED,
		1,
	)

	return comp.Wand.GetImageBlob()
}
