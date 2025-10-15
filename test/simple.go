package test

import (
	"github.com/gofiber/fiber/v2/log"
	"time"
)

func LogTime(f func(), num uint) {
	if num == 0 {
		num = 1
	}
	startTime := time.Now()
	for i := uint(0); i < num; i++ {
		f()
	}
	endTime := time.Now()
	log.Errorf("耗时：%v", endTime.Sub(startTime))
}

func LogTimeWithPreHot(f func(), num uint, preHot ...uint) {
	if num == 0 {
		num = 1
	}
	var prehot uint
	if len(preHot) == 0 {
		prehot = 100000
	} else {
		prehot = preHot[0]
	}
	log.SetLevel(log.LevelFatal)
	for i := uint(0); i < prehot; i++ {
		f()
	}
	log.SetLevel(log.LevelInfo)
	startTime := time.Now()
	for i := uint(0); i < num; i++ {
		f()
	}
	endTime := time.Now()
	log.Errorf("耗时：%v", endTime.Sub(startTime))
}
