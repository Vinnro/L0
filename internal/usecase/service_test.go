package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"L0_main/internal/domain"
	mockCahe "L0_main/internal/infrastructure/cache/mocks"
	mockRep "L0_main/internal/infrastructure/db/mocks"

	"github.com/golang/mock/gomock"
)

func TestInsertOrd_Invalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := mockCahe.NewMockICache(ctrl)
	repoMock := mockRep.NewMockOrder(ctrl)

	invalid := &domain.Order{OrderUID: "1231"}

	repoMock.EXPECT().Insert(invalid).Return(nil)
	cacheMock.EXPECT().Set(invalid.OrderUID, gomock.Any(), gomock.Any()).Return(nil)

	serv := NewSrvice(repoMock, cacheMock, time.Second)
	if err := serv.InsertOrd(invalid); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInsertOrd_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := mockCahe.NewMockICache(ctrl)
	repoMock := mockRep.NewMockOrder(ctrl)

	data, err := os.ReadFile("testdata/order.json")
	if err != nil {
		t.Fatalf("cannot read json: %v", err)
	}

	var ord domain.Order
	if err := json.Unmarshal(data, &ord); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	jsonData, _ := json.Marshal(ord)

	repoMock.EXPECT().Insert(&ord).Return(nil)
	cacheMock.EXPECT().Set(ord.OrderUID, jsonData, gomock.Any()).Return(nil)

	serv := NewSrvice(repoMock, cacheMock, time.Minute)

	if err := serv.InsertOrd(&ord); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCacheOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := mockCahe.NewMockICache(ctrl)
	repoMock := mockRep.NewMockOrder(ctrl)

	ord := &domain.Order{OrderUID: "1231"}
	jsonData, _ := json.Marshal(ord)

	cacheMock.EXPECT().Get("1231").Return(jsonData, nil)
	repoMock.EXPECT().Get(gomock.Any()).Times(0)

	serv := NewSrvice(repoMock, cacheMock, time.Minute)

	res, err := serv.GetOrd("1231")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.OrderUID != "1231" {
		t.Fatalf("wrong uid: got=%v want=1231", res.OrderUID)
	}
}

func TestGetOrdCacheMissDBSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheMock := mockCahe.NewMockICache(ctrl)
	repoMock := mockRep.NewMockOrder(ctrl)

	data, err := os.ReadFile("testdata/order.json")
	if err != nil {
		t.Fatalf("cannot read json: %v", err)
	}

	var ord domain.Order
	if err := json.Unmarshal(data, &ord); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	uid := ord.OrderUID

	cacheMock.EXPECT().
		Get(uid).
		Return(nil, fmt.Errorf("miss"))

	repoMock.EXPECT().
		Get(uid).
		Return(&ord, nil)

	cacheMock.EXPECT().
		Set(uid, gomock.Any(), gomock.Any()).
		Return(nil)

	serv := NewSrvice(repoMock, cacheMock, time.Minute)

	res, err := serv.GetOrd(uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.OrderUID != uid {
		t.Fatalf("wrong uid: %v", res.OrderUID)
	}
}
