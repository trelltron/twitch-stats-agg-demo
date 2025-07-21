package twitch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Cursor string

type Video struct {
	Title    string `json:"title"`
	Views    int    `json:"view_count"`
	Duration string `json:"duration"`
}

type Pagination struct {
	Cursor Cursor `json:"cursor"`
}

type ResponseBody struct {
	Data       []Video    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func (twitch *Service) GetVideos(params url.Values) (*http.Response, error) {
	return twitch.client.get("videos", params)
}

func (twitch *Service) GetUserVideosPage(userId string, limit int, cursor Cursor) ([]Video, Cursor, error) {
	params := make(url.Values)
	params.Add("user_id", userId)
	if limit <= 100 {
		params.Add("first", strconv.FormatUint(uint64(limit), 10))
	} else {
		params.Add("first", "100")
	}

	if len(cursor) > 0 {
		params.Add("after", string(cursor))
	}

	response, err := twitch.GetVideos(params)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		twitch.Log.Debug("Non-success status code recieved", "StatusCode", response.StatusCode, "details", string(body))
		return nil, "", &ApiError{StatusCode: response.StatusCode}
	}

	var data ResponseBody
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		twitch.Log.Error("JSON decoding issue", "err", err)
		return nil, "", err
	}

	return data.Data, data.Pagination.Cursor, nil
}

func (twitch *Service) GetUserVideos(userId string, limit int) ([]Video, error) {
	var (
		batch   []Video
		cursor  Cursor = ""
		err     error
		results []Video
	)

	for {
		batch, cursor, err = twitch.GetUserVideosPage(userId, limit-len(results), cursor)
		if err != nil {
			return results, err
		}
		twitch.Log.Debug("Retrieved page of videos", "count", len(batch), "cursor", cursor)

		results = append(results, batch...)

		if len(results) >= limit || cursor == "" {
			return results, nil
		}
	}
}
