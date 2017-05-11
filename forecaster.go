package main

import (
	"time"
)

// grep "EXT-X-MEDIA-SEQUENCE" luztv-livestream.m3u8
// #EXT-X-MEDIA-SEQUENCE:7173
func diskforecastmechanism() {
	//lastseq := map[string]int{} // lastseq["luztv-livestream.m3u8"] = 7173
	for {
		// 1.- Hacemos un range para recorrer lastseq[] y ver los m3u8 que no existen ya en /live para borrarlos de:
		//     lastseq[], /old, /new

		// 2.- Listar m3u8 de /live y vemos sus media-sequence para determinar cuales actualizar/introducir por primera vez en lastseq[], y cuales no actualizar.
		//     Solo los que actualicen se hará: cp /new /old && cp /live /new (en ese orden)
		//     Los que se introduzcan por primera vez: cp /live /new
		//     Los que no actualizan ni se tocan

		//exec.Command("/bin/sh", "-c", "ls -1 "+rootdir+"/live/*.m3u8 | grep m3u8").Start()

		time.Sleep(100 * time.Millisecond)
	}
}
