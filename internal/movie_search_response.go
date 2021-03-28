package internal

type MovieSearchResponse struct {
	Title    string `json:"Title"`
	Year     string `json:"Year"`
	Rated    string `json:"Rated"`
	Released string `json:"Released"`
	Runtime  string `json:"Runtime"`
	Genre    string `json:"Genre"`
	Director string `json:"Director"`
	Writer   string `json:"Writer"`
	Actors   string `json:"Actors"`
	Plot     string `json:"Plot"`
	Language string `json:"Language"`
	Country  string `json:"Country"`
	Awards   string `json:"Awards"`
	Poster   string `json:"Poster"`
	Ratings  []struct {
		Source string `json:"Source"`
		Value  string `json:"Value"`
	} `json:"Ratings"`
	Metascore    string `json:"Metascore"`
	Imdbrating   string `json:"imdbRating"`
	Imdbvotes    string `json:"imdbVotes"`
	Imdbid       string `json:"imdbID"`
	Type         string `json:"Type"`
	Totalseasons string `json:"totalSeasons"`
	Response     string `json:"Response"`
}
