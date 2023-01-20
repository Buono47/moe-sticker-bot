package core

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func imToWebp(f string) (string, error) {
	pathOut := f + ".webp"
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	args = append(args, "-resize", "512x512", "-filter", "Lanczos", "-define", "webp:lossless=true", f+"[0]", pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("imToWebp ERROR:", string(out))
		return "", err
	}
	return pathOut, err
}

func imToWebpWA(f string) error {
	pathOut := f
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	qualities := []string{"75", "50"}
	for _, q := range qualities {
		args = append(args, "-define", "webp:quality="+q,
			"-resize", "512x512", "-gravity", "center", "-extent", "512x512",
			"-background", "none", f+"[0]", pathOut)

		out, err := exec.Command(bin, args...).CombinedOutput()
		if err != nil {
			log.Warnln("imToWebp ERROR:", string(out))
			return err
		}
		if st, _ := os.Stat(pathOut); st.Size() > 300*KiB {
			continue
		} else {
			return nil
		}
	}
	return errors.New("bad webp")
}

func imToPng(f string) (string, error) {
	pathOut := f + ".png"
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	args = append(args, f, pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("imToPng ERROR:", string(out))
		return "", err
	}
	return pathOut, err
}

func imToGIF(f string) (string, error) {
	pathOut := f + ".gif"
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	args = append(args, f, pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("imToGIF ERROR:", string(out))
		return "", err
	}
	return pathOut, err
}

func ffToWebm(f string) (string, error) {
	pathOut := f + ".webm"
	bin := "ffmpeg"
	args := []string{"-hide_banner", "-i", f,
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
		"-c:v", "libvpx-vp9", "-cpu-used", "5", "-minrate", "50k", "-b:v", "350k", "-maxrate", "450k",
		"-to", "00:00:03", "-an", "-y", pathOut}

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("ffToWebm ERROR:", string(out))
		return pathOut, err
	}

	if stat, _ := os.Stat(pathOut); stat.Size() > 260000 {
		log.Warn("ff to webm too big, retrying...")
		args = []string{"-hide_banner", "-i", f,
			"-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
			"-c:v", "libvpx-vp9", "-cpu-used", "5", "-minrate", "50k", "-b:v", "200k", "-maxrate", "300k",
			"-to", "00:00:03", "-an", "-y", pathOut}
		err = exec.Command(bin, args...).Run()
	}
	return pathOut, err
}

func ffToWebmSafe(f string) (string, error) {
	pathOut := f + ".webm"
	bin := "ffmpeg"
	args := []string{"-hide_banner", "-i", f,
		"-vf", "scale=512:512:force_original_aspect_ratio=decrease:flags=lanczos", "-pix_fmt", "yuva420p",
		"-c:v", "libvpx-vp9", "-cpu-used", "5", "-minrate", "50k", "-b:v", "200k", "-maxrate", "300k",
		"-to", "00:00:02.800", "-r", "30", "-an", "-y", pathOut}

	cmd := exec.Command(bin, args...)
	err := cmd.Run()
	return pathOut, err
}

func ffToGifShrink(f string) (string, error) {
	var decoder []string
	var args []string
	if strings.HasSuffix(f, ".webm") {
		decoder = []string{"-c:v", "libvpx-vp9"}
	}
	pathOut := f + ".gif"
	bin := "ffmpeg"
	args = append(args, decoder...)
	args = append(args, "-i", f, "-hide_banner",
		"-lavfi", "scale=256:256:force_original_aspect_ratio=decrease,split[a][b];[a]palettegen=reserve_transparent=on:transparency_color=ffffff[p];[b][p]paletteuse",
		"-loglevel", "error", "-y", pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("ffToGifShrink ERROR:", string(out))
		return "", err
	}
	return pathOut, err
}

func ffToGif(f string) (string, error) {
	var decoder []string
	var args []string
	if strings.HasSuffix(f, ".webm") {
		decoder = []string{"-c:v", "libvpx-vp9"}
	}
	pathOut := f + ".gif"
	bin := "ffmpeg"
	args = append(args, decoder...)
	args = append(args, "-i", f, "-hide_banner",
		"-lavfi", "split[a][b];[a]palettegen=reserve_transparent=on:transparency_color=ffffff[p];[b][p]paletteuse",
		"-loglevel", "error", "-y", pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnf("ffToGif ERROR:\n%s", string(out))
		return "", err
	}
	return pathOut, err
}

func ffToGifSafe(f string) (string, error) {
	of, err := ffToGif(f)
	if err != nil {
		return "", err
	}
	// GIF should not larget than 20MB
	if stat, _ := os.Stat(of); stat.Size() > 20000000 {
		log.Warn("GIF too big! try shrink")
		of, err = ffToGifShrink(f)
		if err != nil {
			return "", err
		}
	}
	return of, err
}

func imStackToWebp(base string, overlay string) (string, error) {
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	fOut := base + ".composite.webp"

	args = append(args, base, overlay, "-background", "none", "-filter", "Lanczos", "-resize", "512x512", "-composite",
		"-define", "webp:lossless=true", fOut)
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Errorln("IM stack ERROR!", string(out))
		return "", err
	} else {
		return fOut, nil
	}
}

// lottie has severe problem when converting directly to GIF.
// Convert to WEBP first, then GIF.
func lottieToGIF(f string) (string, error) {
	bin := "lottie_convert.py"
	fOut := f + ".webp"
	args := []string{f, fOut}
	out, err := exec.Command(bin, args...).CombinedOutput()
	// fOut, err := ffToGif(fOut)
	if err != nil {
		log.Errorln("lottieToGIF ERROR!", string(out))
		return "", err
	}
	return fOut, nil
}

// Replaces .webm ext to .webp
func imToAnimatedWebpLQ(f string) error {
	pathOut := strings.ReplaceAll(f, ".webm", ".webp")
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	args = append(args, "-resize", "64x64", f, pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("imToWebp ERROR:", string(out))
		return err
	}
	return err
}

// Replaces .webm ext to .webp
// func imToAnimatedWebpWA(f string) error {
// 	pathOut := strings.ReplaceAll(f, ".webm", ".webp")
// 	bin := CONVERT_BIN
// 	args := CONVERT_ARGS

// 	//Try qualities from best to worst.
// 	qualities := []string{"75", "50", "20", "5"}
// 	for _, q := range qualities {
// 		args = append(args,
// 			"-resize", "512x512", "-quality", q,
// 			"-gravity", "center", "-extent", "512x512", "-background", "none",
// 			f, pathOut)

// 		out, err := exec.Command(bin, args...).CombinedOutput()
// 		if err != nil {
// 			log.Warnln("imToWebp ERROR:", string(out))
// 			return err
// 		}
// 		//WhatsApp uses KiB.
// 		if st, _ := os.Stat(pathOut); st.Size() > 300*KiB {
// 			log.Warnf("convert: awebp exceeded 300k, q:%s z:%d s:%s", q, st.Size(), pathOut)
// 			continue
// 		} else {
// 			return nil
// 		}
// 	}
// 	log.Warnln("all retries failed! s:", pathOut)
// 	return errors.New("bad animated webp?")
// }

// // animated webp has a pretty bad compression ratio comparing to VP9,
// // shrink down quality as more as possible.
func ffToAnimatedWebpWA(f string) error {
	pathOut := strings.ReplaceAll(f, ".webm", ".webp")
	bin := "ffmpeg"
	//Try qualities from best to worst.
	qualities := []string{"75", "50", "20", "_DS384"}

	for _, q := range qualities {
		args := []string{"-hide_banner", "-c:v", "libvpx-vp9", "-i", f,
			"-vf", "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:-1:-1:color=black@0",
			"-quality", q, "-loop", "0", "-pix_fmt", "yuva420p",
			"-an", "-y", pathOut}

		if q == "_DS384" {
			args = []string{"-hide_banner", "-c:v", "libvpx-vp9", "-i", f,
				"-vf", "scale=256:256:force_original_aspect_ratio=decrease,pad=512:512:-1:-1:color=black@0",
				"-quality", "20", "-loop", "0", "-pix_fmt", "yuva420p",
				"-an", "-y", pathOut}
		}

		out, err := exec.Command(bin, args...).CombinedOutput()
		if err != nil {
			log.Warnln("ffToAnimatedWebpWA ERROR:", string(out))
			return err
		}
		//WhatsApp uses KiB.
		if st, _ := os.Stat(pathOut); st.Size() > 500*KiB {
			log.Warnf("convert: awebp exceeded 500k, q:%s z:%d s:%s", q, st.Size(), pathOut)
			continue
		} else {
			return nil
		}
	}
	log.Warnln("all quality failed! s:", pathOut)

	return errors.New("bad animated webp?")
}

// Replaces .webm or .webp to .png
func imToPNGThumb(f string) error {
	pathOut := strings.ReplaceAll(f, ".webm", ".png")
	pathOut = strings.ReplaceAll(pathOut, ".webp", ".png")
	bin := CONVERT_BIN
	args := CONVERT_ARGS
	args = append(args,
		"-resize", "96x96",
		"-gravity", "center", "-extent", "96x96", "-background", "none",
		f+"[0]", pathOut)

	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		log.Warnln("imToPng ERROR:", string(out))
		return err
	}
	return err
}
