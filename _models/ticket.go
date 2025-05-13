package models

// Ticket s
type TicketPurchase struct {
	MovieShowID string  `json:"movieShowId" validate:"required,uuid4"`
	SeatID      string  `json:"seatId" validate:"required,uuid4"`
	Price       float64 `json:"price" validate:"required,min=0"`
}

/*
JSON Example:
{
  "movieShowId": "550e8400-e29b-41d4-a716-446655440006",
  "seatId": "550e8400-e29b-41d4-a716-446655440008",
  "price": 15.99
}
*/

type TicketReservation struct {
	MovieShowID string `json:"movieShowId" validate:"required,uuid4"`
	SeatID      string `json:"seatId" validate:"required,uuid4"`
}

/*
JSON Example:
{
  "movieShowId": "550e8400-e29b-41d4-a716-446655440006",
  "seatId": "550e8400-e29b-41d4-a716-446655440008"
}
*/

type TicketResponse struct {
	ID        string            `json:"id"`
	MovieShow MovieShowResponse `json:"movieShow"`
	Seat      SeatResponse      `json:"seat"`
	Status    string            `json:"status"`
	Price     float64           `json:"price"`
}

/*
JSON Example:
{
  "id": "550e8400-e29b-41d4-a716-446655440010",
  "movieShow": {
    "id": "550e8400-e29b-41d4-a716-446655440006",
    "movie": {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "title": "Inception",
      "duration": "02:28:00",
      "rating": 8.8,
      "description": "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
      "ageLimit": 12,
      "boxOfficeRevenue": 836.8,
      "releaseDate": "2010-07-16T00:00:00Z",
      "genres": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440001",
          "name": "Science Fiction",
          "description": "Films that explore futuristic concepts, space travel, time travel, etc."
        },
        {
          "id": "550e8400-e29b-41d4-a716-446655440002",
          "name": "Action",
          "description": "High-energy films with physical stunts, chases, battles, etc."
        }
      ]
    },
    "hall": {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "screen_type": {
        "id": "550e8400-e29b-41d4-a716-446655440004",
        "name": "IMAX",
        "description": "Large-screen film format and projection standard with at least a resolution of 70mm film."
      },
      "name": "Hall 1 - IMAX",
      "capacity": 300,
      "description": "Main hall with IMAX technology and Dolby Atmos sound"
    },
    "startTime": "2023-12-15T18:00:00Z",
    "language": "English",
    "availableSeats": 250
  },
  "seat": {
    "id": "550e8400-e29b-41d4-a716-446655440008",
    "hallId": "550e8400-e29b-41d4-a716-446655440005",
    "seatType": {
      "id": "550e8400-e29b-41d4-a716-446655440007",
      "name": "VIP",
      "description": "Premium seats with extra legroom and comfort"
    },
    "rowNumber": 1,
    "seatNumber": 1
  },
  "status": "Purchased",
  "price": 15.99
}
*/

type TicketUpdate struct {
	Status *string  `json:"status,omitempty" validate:"omitempty,oneof=Purchased Reserved Available"`
	Price  *float64 `json:"price,omitempty" validate:"omitempty,min=0"`
}

/*
JSON Example:
{
  "status": "Reserved",
  "price": 12.99
}
*/
