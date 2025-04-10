package main

type ReservationResult struct {
	Success      bool
	SeatID       int
	PassengerID  int
	Duration     int64 // milisegundos
	Error        error
	Isolation    string
	Concurrency  int
}

type SimulationConfig struct {
	TotalSeats       int
	ConcurrencyLevel int
	IsolationLevel   string
	PassengerIDs     []int
	SeatID          int
}

type SimulationResult struct {
	TotalAttempts   int
	SuccessCount    int
	FailureCount    int
	AvgDuration     float64
	IsolationLevel  string
	Concurrency     int
	Deadlocks       int
}