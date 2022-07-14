//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package query

// MockWorkerPool mock worker pool implementation for tests
type MockWorkerPool struct {
	resp     *HTTPResponse
	checkReq func(*HTTPRequest)
}

func NewMockWorkerPool() *MockWorkerPool {
	return &MockWorkerPool{}
}

func (mwp *MockWorkerPool) MockResp(resp *HTTPResponse) {
	mwp.resp = resp
}

func (mwp *MockWorkerPool) MockBody(body string) {
	mwp.resp = &HTTPResponse{
		Body: []byte(body),
	}
}
func (mwp *MockWorkerPool) SetRequestAssertions(f func(*HTTPRequest)) {
	mwp.checkReq = f
}

func (mwp *MockWorkerPool) Query(req *HTTPRequest) *HTTPResponse {
	if mwp.checkReq != nil {
		mwp.checkReq(req)
	}
	return mwp.resp
}
