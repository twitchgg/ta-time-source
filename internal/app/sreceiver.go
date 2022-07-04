package app

func (tsa *TimeSourceApp) _startSReceiver(errChan chan error) {
	go func() {
		if err := <-tsa.gpReceiver.Open("$GPRMC"); err != nil {
			errChan <- err
			return
		}
	}()
	go func() {
		if err := <-tsa.gbReceiver.Open("$GBRMC"); err != nil {
			errChan <- err
			return
		}
	}()
	tsa.gpReceiver.ReadTime(errChan)
	tsa.gbReceiver.ReadTime(errChan)
}
