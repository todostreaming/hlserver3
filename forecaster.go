package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// grep "EXT-X-MEDIA-SEQUENCE" luztv-livestream.m3u8
// #EXT-X-MEDIA-SEQUENCE:7173
func diskforecastmechanism() {
	lastseq := map[string]int{} // lastseq["luztv-livestream.m3u8"] = 7173
	var err error

	for {
		// 1.- Hacemos un range para recorrer lastseq[] y ver los m3u8 que no existen ya en /live para borrarlos de:
		//     lastseq[], /old, /new
		for k, _ := range lastseq {
			if _, err = os.Stat(rootdir + "live/" + k); os.IsNotExist(err) {
				delete(lastseq, k)
				exec.Command("/bin/sh", "-c", "rm "+rootdir+"live/old/"+k).Run()
				exec.Command("/bin/sh", "-c", "rm "+rootdir+"live/new/"+k).Run()
			}
		}
		// 2.- Listar m3u8 de /live y vemos sus media-sequence para determinar cuales actualizar/introducir por primera vez, y cuales no actualizar.
		//     Solo los que actualicen se hará: cp /new /old && cp /live /new (en ese orden)
		//     Los que se introduzcan por primera vez: cp /live /new
		//     Los que no actualizan ni se tocan
		list := []string{} // list of m3u8 in /live
		cmd := exec.Command("/bin/sh", "-c", "ls -1 "+rootdir+"/live/ | grep .m3u8")
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
			val, ok := lastseq[v]
			if !ok { // es  nuevo hay q darlo de alta
				lastseq[v] = msq
				exec.Command("/bin/sh", "-c", "cp "+rootdir+"live/"+v+" "+rootdir+"new/").Run()
			} else {
				if msq != val {
					lastseq[v] = msq
					exec.Command("/bin/sh", "-c", "cp "+rootdir+"new/"+v+" "+rootdir+"old/").Run()
					exec.Command("/bin/sh", "-c", "cp "+rootdir+"live/"+v+" "+rootdir+"new/").Run()
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
