package export

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"math"
	"strings"

	"shapegen/render"
)

// ── COCO JSON structures ───────────────────────────────────────────────────────

// CocoDataset is the top-level COCO format object.
type CocoDataset struct {
	Info        CocoInfo         `json:"info"`
	Categories  []CocoCategory   `json:"categories"`
	Images      []CocoImage      `json:"images"`
	Annotations []CocoAnnotation `json:"annotations"`
}

// CocoInfo holds dataset metadata.
type CocoInfo struct {
	Description string `json:"description"`
	Version     string `json:"version"`
}

// CocoCategory describes a single object category.
type CocoCategory struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Supercategory string `json:"supercategory"`
}

// CocoImage describes a single image entry.
type CocoImage struct {
	ID       int    `json:"id"`
	FileName string `json:"file_name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// CocoAnnotation describes a single bounding-box annotation.
type CocoAnnotation struct {
	ID         int        `json:"id"`
	ImageID    int        `json:"image_id"`
	CategoryID int        `json:"category_id"`
	BBox       [4]float64 `json:"bbox"`
	Area       float64    `json:"area"`
	IsCrowd    int        `json:"iscrowd"`
}

// ── VOC XML structures ─────────────────────────────────────────────────────────

// VocAnnotation is the root Pascal VOC XML element.
type VocAnnotation struct {
	XMLName  xml.Name  `xml:"annotation"`
	Filename string    `xml:"filename"`
	Size     VocSize   `xml:"size"`
	Object   VocObject `xml:"object"`
}

// VocSize holds image dimensions.
type VocSize struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
	Depth  int `xml:"depth"`
}

// VocObject describes the annotated object.
type VocObject struct {
	Name   string    `xml:"name"`
	BndBox VocBndBox `xml:"bndbox"`
}

// VocBndBox holds the pixel coordinates of the bounding box.
type VocBndBox struct {
	Xmin int `xml:"xmin"`
	Ymin int `xml:"ymin"`
	Xmax int `xml:"xmax"`
	Ymax int `xml:"ymax"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// EncodePNG encodes an image.Image to PNG bytes.
func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func addToZip(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// ── Format writers ────────────────────────────────────────────────────────────

// WritePlain writes images organised by class directory (no annotation files).
func WritePlain(zw *zip.Writer, dataset []render.ClassImages) error {
	for _, cls := range dataset {
		for i, img := range cls.Images {
			name := fmt.Sprintf("%s/%s_%04d.png", cls.Class.Name, cls.Class.Name, i)
			if err := addToZip(zw, name, img.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteCOCO writes images and a single COCO-format JSON annotation file.
func WriteCOCO(zw *zip.Writer, dataset []render.ClassImages, size int) error {
	coco := CocoDataset{
		Info: CocoInfo{Description: "2D Shape Dataset", Version: "1.0"},
	}
	for _, cls := range dataset {
		coco.Categories = append(coco.Categories, CocoCategory{
			ID: cls.Class.ID + 1, Name: cls.Class.Name, Supercategory: "shape",
		})
	}
	imgID, annID := 1, 1
	for _, cls := range dataset {
		for i, img := range cls.Images {
			fname := fmt.Sprintf("images/%s_%04d.png", cls.Class.Name, i)
			coco.Images = append(coco.Images, CocoImage{
				ID: imgID, FileName: fname, Width: size, Height: size,
			})
			bb := img.BBox
			coco.Annotations = append(coco.Annotations, CocoAnnotation{
				ID: annID, ImageID: imgID, CategoryID: cls.Class.ID + 1,
				BBox:    [4]float64{round2(bb.X), round2(bb.Y), round2(bb.W), round2(bb.H)},
				Area:    round2(bb.W * bb.H),
				IsCrowd: 0,
			})
			if err := addToZip(zw, fname, img.Data); err != nil {
				return err
			}
			imgID++
			annID++
		}
	}
	jsonBytes, err := json.MarshalIndent(coco, "", "  ")
	if err != nil {
		return err
	}
	return addToZip(zw, "annotations/instances.json", jsonBytes)
}

// WriteYOLO writes images, per-image label files, and a data.yaml descriptor.
func WriteYOLO(zw *zip.Writer, dataset []render.ClassImages, size int) error {
	var sb strings.Builder
	sb.WriteString("path: .\ntrain: images\nval: images\n")
	sb.WriteString(fmt.Sprintf("nc: %d\n", len(dataset)))
	names := make([]string, len(dataset))
	for i, cls := range dataset {
		names[i] = fmt.Sprintf("'%s'", cls.Class.Name)
	}
	sb.WriteString(fmt.Sprintf("names: [%s]\n", strings.Join(names, ", ")))
	if err := addToZip(zw, "data.yaml", []byte(sb.String())); err != nil {
		return err
	}
	for _, cls := range dataset {
		for i, img := range cls.Images {
			base := fmt.Sprintf("%s_%04d", cls.Class.Name, i)
			if err := addToZip(zw, "images/"+base+".png", img.Data); err != nil {
				return err
			}
			bb := img.BBox
			f := float64(size)
			cx := (bb.X + bb.W/2) / f
			cy := (bb.Y + bb.H/2) / f
			nw := bb.W / f
			nh := bb.H / f
			label := fmt.Sprintf("%d %.6f %.6f %.6f %.6f\n", cls.Class.ID, cx, cy, nw, nh)
			if err := addToZip(zw, "labels/"+base+".txt", []byte(label)); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteVOC writes images and per-image Pascal VOC XML annotation files.
func WriteVOC(zw *zip.Writer, dataset []render.ClassImages, size int) error {
	for _, cls := range dataset {
		for i, img := range cls.Images {
			base := fmt.Sprintf("%s_%04d", cls.Class.Name, i)
			if err := addToZip(zw, "JPEGImages/"+base+".png", img.Data); err != nil {
				return err
			}
			bb := img.BBox
			ann := VocAnnotation{
				Filename: base + ".png",
				Size:     VocSize{Width: size, Height: size, Depth: 3},
				Object: VocObject{
					Name: cls.Class.Name,
					BndBox: VocBndBox{
						Xmin: int(math.Round(bb.X)),
						Ymin: int(math.Round(bb.Y)),
						Xmax: int(math.Round(bb.X + bb.W)),
						Ymax: int(math.Round(bb.Y + bb.H)),
					},
				},
			}
			xmlBytes, err := xml.MarshalIndent(ann, "", "  ")
			if err != nil {
				return err
			}
			xmlBytes = append([]byte(xml.Header), xmlBytes...)
			if err := addToZip(zw, "Annotations/"+base+".xml", xmlBytes); err != nil {
				return err
			}
		}
	}
	return nil
}
