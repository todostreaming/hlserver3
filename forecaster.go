package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// /var/segments/live = will contain the realtime m3u8 playlists and .ts segments
// /var/segments/new = will contain the new playlist with the latest segment to pre-fetch in cache
// /var/segments/old = will contain the real playlist to be played with .ts already pre-fetched in caches (no need to access backend)
func diskforecastmechanism() {
	lastseq := map[string]int{} // lastseq["luztv-livestream.m3u8"] = 7173
	var err error

	for {
		// 1.- Hacemos un range para recorrer lastseq[] y ver los m3u8 que no existen ya en /live para borrarlos de:
		//     lastseq[], /old, /new
		for k, _ := range lastseq {
			if _, err = os.Stat(rootdir + "live/" + k); os.IsNotExist(err) {
				delete(lastseq, k)
				str1 := fmt.Sprintf("rm %sold/%s", rootdir, k)
				str2 := fmt.Sprintf("rm %snew/%s", rootdir, k)
				exec.Command("/bin/sh", "-c", str1).Run()
				exec.Command("/bin/sh", "-c", str2).Run()
			}
		}
		// 2.- Listar m3u8 de /live y vemos sus media-sequence para determinar cuales actualizar/introducir por primera vez, y cuales no actualizar.
		//     Solo los que actualicen se hará: cp /new /old && cp /live /new (en ese orden)
		//     Los que se introduzcan por primera vez: cp /live /new
		//     Los que no actualizan ni se tocan
		list := []string{} // list of m3u8 in /live
		str3 := fmt.Sprintf("ls -1 %slive/ | grep .m3u8", rootdir)
		cmd := exec.Command("/bin/sh", "-c", str3)
		stdoutRead, _ := cmd.StdoutPipe()
		reader := bufio.NewReader(stdoutRead)
		cmd.Start()
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\n")
			list = append(list, line)
		}
		cmd.Wait()
		// ya tenemos el listado de ficheros m3u8 en /live ([]list), lo recorremos completo
		for _, v := range list {
			msq := getmediaseqnum(rootdir + "live/" + v)
			if msq == 0 {
				continue
			}
			val, ok := lastseq[v]
			if !ok { // es  nuevo hay q darlo de alta
				lastseq[v] = msq
				str4 := fmt.Sprintf("cp %slive/%s %snew/", rootdir, v, rootdir)
				exec.Command("/bin/sh", "-c", str4).Run()
			} else {
				if msq != val {
					lastseq[v] = msq
					str5 := fmt.Sprintf("cp %snew/%s %sold/", rootdir, v, rootdir)
					str6 := fmt.Sprintf("cp %slive/%s %snew/", rootdir, v, rootdir)
					exec.Command("/bin/sh", "-c", str5).Run()
					exec.Command("/bin/sh", "-c", str6).Run()
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func getmediaseqnum(m3u8 string) int {
	var seq int

	out, err := exec.Command("/bin/sh", "-c", "grep \"EXT-X-MEDIA-SEQUENCE\" "+m3u8).CombinedOutput()
	if err == nil {
		fmt.Sscanf(string(out), "#EXT-X-MEDIA-SEQUENCE:%d\n", &seq)
	}

	return seq
}

func getlatestseg(rawstream string) string { // tail -1 /var/segments/new/luztv-livestream.m3u8
	var seg string

	out, err := exec.Command("/bin/sh", "-c", "tail -1 "+rootdir+"new/"+rawstream+".m3u8").CombinedOutput()
	if err == nil {
		seg = strings.TrimRight(string(out), "\n")
	}

	return seg
}
