package litesockets_test

import (
	"testing"
	"time"

	"github.com/gabe-lee/litesockets"
)

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i += 1 {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func FuzzSocket(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{})
	f.Add([]byte(nil))
	f.Add([]byte(lorem))
	f.Add(rainbow)
	loopErrors := make(chan error, 20)
	server, errS := litesockets.NewSimpleSocketServer("127.0.0.1:5555", 10, time.Second*10, func(socket *litesockets.Socket) {
		for {
			msg, errSR := socket.Read()
			if errSR != nil {
				loopErrors <- errSR
			}
			_, errSW := socket.Write(msg)
			if errSW != nil {
				loopErrors <- errSR
			}
		}
	})
	if errS != nil {
		loopErrors <- errS
	}
	go server.BeginServing()
	go func() {
		e := <-server.Errors
		loopErrors <- e
	}()
	client, errC := litesockets.OpenSocket("127.0.0.1:5555", time.Second*10)
	if errC != nil {
		loopErrors <- errC
	}
	f.Fuzz(func(t *testing.T, a []byte) {
		_, errW := client.Write(a)
		b, errR := client.Read()
		var errL error
		select {
		case e := <-loopErrors:
			errL = e
		default:
			errL = nil
		}
		if errL != nil {
			t.Errorf("Loop Error: %s", errL.Error())
		}
		if errW != nil || errR != nil {
			t.Error("Client Read/Write produced errors", errW, errR)
		}
		if !bytesEqual(a, b) {
			t.Error("Client Read/Write didn't match")
		}
	})
}

var rainbow = func() []byte {
	count := 100000000
	val := byte(0)
	data := make([]byte, count)
	for i := 0; i < count; i += 1 {
		data[i] = val
		val = val + 1
	}
	return data
}()

const lorem = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce nisi erat, imperdiet at faucibus vitae, venenatis consectetur mi. Duis convallis neque in diam rhoncus tristique eu quis nulla. Praesent mattis arcu elit, finibus lacinia sem vehicula ullamcorper. Etiam vel ante a mauris egestas pulvinar ullamcorper ac sem. Curabitur gravida urna eu diam accumsan ultricies. Nunc tempus justo mi, eu faucibus mauris dictum id. Cras a ultrices quam. Curabitur id dignissim leo, sit amet rhoncus dui. Maecenas tincidunt interdum augue, sed fermentum sem ultrices vel. Proin ac hendrerit felis. Mauris eu massa sed libero tincidunt dignissim in ut dui.

Vestibulum orci ligula, aliquet eu mi ut, commodo accumsan justo. Phasellus volutpat fermentum tempus. Nunc ac auctor tellus. Vivamus interdum metus urna, eu vulputate diam aliquet sit amet. Nullam tristique fermentum nunc eget sodales. Vestibulum consectetur dolor vitae varius dictum. Donec dignissim dolor leo, consectetur rutrum purus tincidunt euismod. Cras congue ullamcorper sapien, a auctor lorem volutpat ac. Sed varius porta enim ac egestas. Duis imperdiet quis tortor ac pellentesque.

Donec hendrerit fringilla sapien, id suscipit lacus facilisis nec. Nullam ut lorem et leo rutrum ullamcorper. Fusce a dui posuere, volutpat diam ac, viverra mi. Nulla molestie commodo finibus. Duis scelerisque euismod elit, sit amet mattis dolor vehicula et. Ut pretium tellus nec turpis hendrerit lacinia. Nulla semper rutrum odio. Quisque ut ligula nec massa sodales lobortis. Nulla varius volutpat tempor. Nullam erat sem, euismod id sem vitae, tempus ultricies nulla. Suspendisse potenti. Donec ut nisi volutpat, sollicitudin quam a, hendrerit sem. Pellentesque imperdiet enim id turpis accumsan placerat sit amet ut sapien. Donec sed vulputate velit. Cras ullamcorper est et ligula commodo, nec eleifend velit tempus.

In sodales justo eget mauris semper accumsan. Vivamus eget rutrum ipsum. Vivamus maximus turpis quis diam efficitur vulputate. Ut fermentum urna et dolor malesuada, facilisis accumsan sapien euismod. Pellentesque sapien purus, porttitor at imperdiet nec, laoreet a dui. Sed vehicula sodales nisl sed mattis. Cras mauris mauris, fermentum id condimentum sit amet, accumsan at dui. Nulla facilisi. Aliquam odio nulla, convallis et tristique ullamcorper, hendrerit vehicula ligula. Interdum et malesuada fames ac ante ipsum primis in faucibus.

Phasellus quis odio lorem. Proin sagittis lacus nec lacus dapibus varius. Integer mattis purus id elementum scelerisque. Quisque efficitur commodo ligula, et ullamcorper ex imperdiet ut. Morbi a libero a sem commodo facilisis sit amet eu nisl. Fusce a augue ut nisi aliquam pulvinar at non augue. Nulla metus elit, iaculis in enim ut, egestas pulvinar nisl. Nunc et nunc non nunc pretium tincidunt a ullamcorper ipsum. Suspendisse potenti. Mauris vel nisi tempus, venenatis lacus eu, vulputate libero. Cras semper semper elit quis suscipit. Nullam et ligula tincidunt nunc pellentesque efficitur sit amet ultrices magna. Maecenas hendrerit, eros finibus pharetra porta, mauris arcu ornare mauris, vitae posuere nibh nisl suscipit nisl. Mauris ut dignissim risus, id porta metus. Vivamus feugiat ante eget urna maximus laoreet.`
