package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leandrofars/oktopus/internal/bridge"
	"github.com/leandrofars/oktopus/internal/entity"
	local "github.com/leandrofars/oktopus/internal/nats"
	"github.com/leandrofars/oktopus/internal/usp/usp_msg"
	"github.com/leandrofars/oktopus/internal/usp/usp_record"
	"github.com/leandrofars/oktopus/internal/usp/usp_utils"
	"github.com/leandrofars/oktopus/internal/utils"
	"google.golang.org/protobuf/proto"
)

func (a *Api) deviceGetSupportedParametersMsg(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn := vars["sn"]

	device, err := bridge.NatsReq[entity.Device](
		local.NATS_ADAPTER_SUBJECT+sn+".device",
		[]byte(""),
		w,
		a.nc,
	)
	if err != nil {
		return
	}

	if device.Msg.Status != entity.Online {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(utils.Marshall("Device is offline"))
		return
	}

	var getSupportedDM usp_msg.GetSupportedDM
	utils.MarshallDecoder(&getSupportedDM, r.Body)
	msg := usp_utils.NewGetSupportedParametersMsg(getSupportedDM)
	protoMsg, err := proto.Marshal(&msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(utils.Marshall(err.Error()))
		return
	}
	record := usp_utils.NewUspRecord(protoMsg, sn)
	protoRecord, err := proto.Marshal(&record)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(utils.Marshall(err.Error()))
		return
	}

	data, err := bridge.NatsUspInteraction(
		local.NATS_MQTT_SUBJECT_PREFIX+sn+".api",
		local.NATS_MQTT_ADAPTER_SUBJECT_PREFIX+sn+".api",
		protoRecord,
		w,
		a.nc,
	)
	if err != nil {
		return
	}

	var receivedRecord usp_record.Record
	err = proto.Unmarshal(data, &receivedRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(utils.Marshall(err.Error()))
		return
	}
	var receivedMsg usp_msg.Msg
	err = proto.Unmarshal(receivedRecord.GetNoSessionContext().Payload, &receivedMsg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(utils.Marshall(err.Error()))
		return
	}
	body := receivedMsg.Body.GetResponse()
	utils.MarshallEncoder(body.GetGetSupportedDmResp(), w)
}

// func (a *Api) retrieveDevices(w http.ResponseWriter, r *http.Request) {
// 	const PAGE_SIZE_LIMIT = 50
// 	const PAGE_SIZE_DEFAULT = 20

// 	// Get specific device
// 	id := r.URL.Query().Get("id")
// 	if id != "" {
// 		device, err := a.Db.RetrieveDevice(id)
// 		if err != nil {
// 			if err == mongo.ErrNoDocuments {
// 				json.NewEncoder(w).Encode("Device id: " + id + " not found")
// 				return
// 			}
// 			json.NewEncoder(w).Encode(err)
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		err = json.NewEncoder(w).Encode(device)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 		return
// 	}

// 	// Get devices with pagination
// 	page_n := r.URL.Query().Get("page_number")
// 	page_s := r.URL.Query().Get("page_size")
// 	var err error

// 	var page_number int64
// 	if page_n == "" {
// 		page_number = 0
// 	} else {
// 		page_number, err = strconv.ParseInt(page_n, 10, 64)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			json.NewEncoder(w).Encode("Page number must be an integer")
// 			return
// 		}
// 	}

// 	var page_size int64
// 	if page_s != "" {
// 		page_size, err = strconv.ParseInt(page_s, 10, 64)

// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			json.NewEncoder(w).Encode("Page size must be an integer")
// 			return
// 		}

// 		if page_size > PAGE_SIZE_LIMIT {
// 			w.WriteHeader(http.StatusBadRequest)
// 			json.NewEncoder(w).Encode("Page size must not exceed " + strconv.Itoa(PAGE_SIZE_LIMIT))
// 			return
// 		}

// 	} else {
// 		page_size = PAGE_SIZE_DEFAULT
// 	}

// 	total, err := a.Db.RetrieveDevicesCount(bson.M{})
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode("Unable to get devices count from database")
// 		return
// 	}

// 	skip := page_number * (page_size - 1)
// 	if total < page_size {
// 		skip = 0
// 	}

// 	//TODO: Create filters
// 	//TODO: Create sorting

// 	sort := bson.M{}
// 	sort["status"] = 1

// 	filter := bson.A{
// 		//bson.M{"$match": filter},
// 		bson.M{"$sort": sort}, // shows online devices first
// 		bson.M{"$skip": skip},
// 		bson.M{"$limit": page_size},
// 	}

// 	devices, err := a.Db.RetrieveDevices(filter)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode("Unable to aggregate database devices info")
// 		return
// 	}

// 	err = json.NewEncoder(w).Encode(map[string]interface{}{
// 		"pages":   total / page_size,
// 		"page":    page_number,
// 		"size":    page_size,
// 		"devices": devices,
// 	})
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

// func (a *Api) deviceCreateMsg(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]
// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.Add

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewCreateMsg(receiver)
// 	a.uspCall(msg, sn, w, device)
// }

// func (a *Api) deviceGetMsg(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]

// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.Get

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewGetMsg(receiver)
// 	a.uspCall(msg, sn, w, device)
// }

// func (a *Api) deviceOperateMsg(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]

// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.Operate

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewOperateMsg(receiver)
// 	a.uspCall(msg, sn, w, device)
// }

// func (a *Api) deviceDeleteMsg(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]
// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.Delete

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewDelMsg(receiver)
// 	a.uspCall(msg, sn, w, device)

// 	//a.Broker.Request(tr369Message, usp_msg.Header_GET, "oktopus/v1/agent/"+sn, "oktopus/v1/get/"+sn)

// }

// func (a *Api) deviceUpdateMsg(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]
// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.Set

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewSetMsg(receiver)
// 	a.uspCall(msg, sn, w, device)
// }

// // TODO: react this function, return err and deal with it in the caller, remove header superfluos
// func (a *Api) deviceExists(sn string, w http.ResponseWriter) db.Device {
// 	device, err := a.Db.RetrieveDevice(sn)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			w.WriteHeader(http.StatusBadRequest)
// 			json.NewEncoder(w).Encode("No device with serial number " + sn + " was found")
// 		}
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return device
// 	}
// 	return device
// }

// func (a *Api) deviceGetParameterInstances(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	sn := vars["sn"]
// 	device := a.deviceExists(sn, w)

// 	var receiver usp_msg.GetInstances

// 	err := json.NewDecoder(r.Body).Decode(&receiver)
// 	if err != nil {
// 		log.Println(err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	msg := utils.NewGetParametersInstancesMsg(receiver)
// 	a.uspCall(msg, sn, w, device)
// }
