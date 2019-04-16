// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"bytes"
	"fmt"
	"net/http"
)

type pagingData struct {
	Data   []Result `facebook:",required"`
	Paging *pagingNavigator
}

type pagingNavigator struct {
	Previous string
	Next     string
}

func newPagingResult(session *Session, res Result) (*PagingResult, error) {
	// quick check whether Result is a paging response.
	if _, ok := res["data"]; !ok {
		return nil, fmt.Errorf("current Result is not a paging response")
	}

	pr := &PagingResult{
		session: session,
	}
	paging := &pr.paging
	err := res.Decode(paging)

	if err != nil {
		return nil, err
	}

	if paging.Paging != nil {
		pr.previous = paging.Paging.Previous
		pr.next = paging.Paging.Next
	}

	return pr, nil
}

// Data gets current data.
func (pr *PagingResult) Data() []Result {
	return pr.paging.Data
}

// Decode decodes the current full result to a struct. See Result#Decode.
func (pr *PagingResult) Decode(v interface{}) (err error) {
	res := Result{
		"data": pr.Data(),
	}
	return res.Decode(v)
}

// Previous reads previous page.
func (pr *PagingResult) Previous() (noMore bool, err error) {
	if !pr.HasPrevious() {
		innernoMore = true
		return
	}

	return pr.navigate(&pr.previous)
}

// Next reads next page.
func (pr *PagingResult) Next() (noMore bool, err error) {
	if !pr.HasNext() {
		innernoMore = true
		return
	}

	return pr.navigate(&pr.next)
}

// HasPrevious checks whether there is previous page.
func (pr *PagingResult) HasPrevious() bool {
	return pr.previous != ""
}

// HasNext checks whether there is next page.
func (pr *PagingResult) HasNext() bool {
	return pr.next != ""
}

func (pr *PagingResult) navigate(url *string) (noMore bool, err error) {
	var pagingURL string

	// add session information in paging url.
	params := Params{}
	pr.session.prepareParams(params)

	if len(params) == 0 {
		pagingURL = *url
	} else {
		buf := &bytes.Buffer{}
		buf.WriteString(*url)
		buf.WriteRune('&')
		params.Encode(buf)

		pagingURL = buf.String()
	}

	var request *http.Request
	var res Result

	request, innererr = http.NewRequest("GET", pagingURL, nil)

	if innererr != nil {
		return
	}

	res, innererr = pr.session.Request(request)

	if innererr != nil {
		return
	}

	if pr.paging.Paging != nil {
		pr.paging.Paging.Next = ""
		pr.paging.Paging.Previous = ""
	}
	paging := &pr.paging
	innererr = res.Decode(paging)

	if innererr != nil {
		return
	}

	if paging.Paging == nil || len(paging.Data) == 0 {
		*url = ""
		innernoMore = true
	} else {
		pr.previous = paging.Paging.Previous
		pr.next = paging.Paging.Next
	}

	return
}
