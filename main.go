// package main contains all processes for this application.
//
// # Polling and not long polling or alternatives
//
// Using enddpoints with longpolling is be beneficial because the client is
// likely to request data more frequently. The rrecent http/3 standard uses UDP
// and with a smaller overhead for each connection. The compatability of regular
// http requests are unparallelled.
//
// For longpolling to work, the client needs to call the API asyncronously and
// noot tie up the main thread, and also needs to return the timestamp of the
// last request they recieved to properly assess when they shuld get the next
// one.
//
// App Engine application can have a maximum of 100 paralell connections to each
// instance of the application. When using autoscaling the maximum lifespan of a
// connection is 10 minutes. The Least expensive F1 instance scales automaticly
// when latency is high, and this is a huge problem since two instances results
// higher cost.
//
// However, since the requests are so small and light on resources combined with
// regular updates there are no real beneffits to long polling. The overhead for
// each send/recieve transaction is about 900 bytes with 6 cookies for Google
// Analytics, GA4 and Cloudflares "allways online" cookie included.
//
//	rtgeo     h3-29  6 kb
//	rtuserss  h3-29  25 bytes
//	rtdevices h3-29  52 bytes
//
// Each request is somewhat slow on the app engine with a 100ms roundtrip from
// Oslo via cloudflare to google cloud.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/analytics/v3"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const (
	AppPrefix = "/dashboard/"
)

var db sync.Map

const (
	RTUsers         string        = "rtusers"
	RTGeo           string        = "rtgeo"
	RTDevices       string        = "rtdevices"
	LastModified    string        = "lastmodified"
	updateFrequency time.Duration = 30 * time.Second
)

func main() {

	// Initialize random
	rand.Seed(time.Now().UnixNano())

	var err error
	err = RefreshData(context.Background())
	if err != nil {
		log.Println(err)
	}

	// Set up all endpoints.
	http.HandleFunc(AppPrefix+"rtusers", RtUsersHandler)
	http.HandleFunc(AppPrefix+"rtgeo", RtGeoHandler)
	http.HandleFunc(AppPrefix+"rtdevices", RtDeviceHandler)
	http.HandleFunc(AppPrefix+"refreshdata", RefreshDataHandler)
	http.Handle(AppPrefix, http.StripPrefix(AppPrefix, http.FileServer(http.Dir("public"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
func RtUsersHandler(w http.ResponseWriter, r *http.Request) {
	setCacheHeadersForHTTPHandlers(w)

	var jsonResponse string
	value, ok := db.Load(RTUsers)
	if !ok {
		jsonResponse = `{"rt:activeUsers":"0"}`
	} else {
		arr := value.([][]string)
		jsonResponse = fmt.Sprintf("{\"rt:activeUsers\":\"%v\"}", arr[0][0])
	}
	fmt.Fprint(w, jsonResponse)
}

// RtGeoHandler serves a array with geo-coordinates and the number of active
// sessions.
func RtGeoHandler(w http.ResponseWriter, r *http.Request) {
	setCacheHeadersForHTTPHandlers(w)

	var jsonResponse string
	value, ok := db.Load(RTGeo)
	jsonResponse = `[["-0.080784","53.567471","2"],["-0.420025","51.878674","1"],["-0.973177","51.150719","1"],["-118.243683","34.052235","2"],["-123.120743","49.282730","4"],["-3.703790","40.416775","1"],["-46.633308","-23.550520","5"],["-5.984459","37.389091","1"],["-7.589843","33.573109","1"],["-73.567261","45.501690","1"],["-74.005974","40.712776","7"],["-74.072090","4.710989","1"],["-77.373314","37.608757","1"],["0.000000","0.000000","150"],["10.011667","59.055801","12"],["10.043347","63.274067","1"],["10.122765","54.323288","1"],["10.148064","59.703354","32"],["10.181531","36.806496","2"],["10.209221","59.582016","13"],["10.215821","59.132153","15"],["10.222798","59.884216","8"],["10.224922","60.703938","1"],["10.303763","59.364986","14"],["10.329676","63.158043","5"],["10.383065","63.368736","67"],["10.428460","59.751186","16"],["10.448479","60.240685","1"],["10.463612","61.113529","12"],["10.465049","59.398579","11"],["10.478613","59.957420","31"],["10.500107","60.894142","7"],["10.500222","61.342010","1"],["10.572784","60.359200","5"],["10.583865","60.306759","1"],["10.625467","60.648968","4"],["10.638197","59.804283","3"],["10.671517","59.708134","2"],["10.685109","59.567013","2"],["10.698507","59.426048","15"],["10.735163","60.908985","14"],["10.752410","59.914181","396"],["10.776710","61.603813","1"],["10.793012","59.686432","2"],["10.849522","60.114437","6"],["10.872416","59.351711","3"],["10.872568","59.062469","2"],["10.911690","63.466995","4"],["10.930987","59.220383","23"],["10.954183","59.923801","10"],["10.956988","60.269619","7"],["103.867744","1.355379","6"],["11.009578","60.089371","1"],["11.041494","63.844994","2"],["11.073340","59.767994","1"],["11.073385","60.792831","3"],["11.089188","55.686432","1"],["11.109212","59.284718","15"],["11.163789","60.134911","6"],["11.193516","60.720230","10"],["11.193996","63.187317","1"],["11.272252","60.329643","4"],["11.300446","63.746426","4"],["11.336381","66.107109","1"],["11.336404","64.517693","2"],["11.350733","60.871464","1"],["11.384262","62.574936","3"],["11.393724","59.114094","3"],["11.479550","60.154533","4"],["11.577461","59.794266","4"],["11.613433","64.015495","10"],["11.621586","59.486309","1"],["11.696552","60.205780","5"],["11.706567","59.268551","1"],["11.746684","60.955086","4"],["11.916895","63.750599","5"],["11.974560","57.708870","4"],["11.994639","60.192524","6"],["12.010858","60.613197","3"],["12.087845","55.641907","1"],["12.132717","59.389160","1"],["12.245401","65.505020","4"],["12.342978","55.868473","1"],["12.460837","65.874077","2"],["12.568336","55.676098","9"],["12.592136","59.654854","1"],["12.857789","56.674374","1"],["13.046525","67.900200","1"],["13.200150","65.811203","4"],["13.511497","59.402184","1"],["13.734843","68.227638","2"],["14.312232","66.439339","11"],["14.415077","67.282921","7"],["14.443382","68.232216","4"],["14.552812","53.428543","2"],["14.848283","67.755760","2"],["15.000448","68.524200","2"],["15.391778","67.258896","1"],["15.412724","68.706276","4"],["15.593840","69.253189","3"],["15.620805","68.319191","1"],["15.849404","67.512177","2"],["153.025131","-27.469770","6"],["16.373819","48.208172","1"],["16.581938","68.912605","9"],["16.871872","41.117142","1"],["16.925167","52.406376","1"],["17.038538","51.107883","1"],["17.506233","68.274109","2"],["17.806826","51.654987","1"],["18.068581","59.329327","10"],["18.197201","54.334297","1"],["18.251074","52.223034","1"],["18.671381","50.294495","1"],["18.984121","69.649773","37"],["19.023781","50.264893","1"],["19.260958","69.217117","1"],["19.455984","51.759247","1"],["2.259290","48.900551","1"],["20.227049","69.201767","1"],["20.619795","52.881485","1"],["20.973364","70.033920","1"],["20.985840","50.012104","1"],["21.012228","52.229675","5"],["21.389990","69.557228","3"],["22.568447","51.246456","2"],["23.271587","69.968880","3"],["23.321867","42.697708","1"],["23.571709","70.530800","2"],["24.357477","55.734791","1"],["24.753574","59.436958","6"],["24.938379","60.169861","3"],["25.279652","54.687157","1"],["25.301447","70.385155","2"],["25.783165","71.169495","1"],["26.102537","44.426769","1"],["29.685946","69.757858","5"],["29.811670","70.158333","2"],["3.554402","51.482395","1"],["30.523399","50.450100","6"],["31.208851","30.013056","1"],["31.441282","36.786869","1"],["33.022617","34.707130","1"],["39.208328","-6.792354","1"],["4.351710","50.850338","1"],["4.477732","51.924419","2"],["4.759839","60.404301","9"],["4.871985","50.467388","6"],["4.894540","52.366699","8"],["5.089330","60.461876","4"],["5.181769","59.795033","4"],["5.206139","58.861759","2"],["5.276590","59.413799","9"],["5.280176","59.282303","15"],["5.317035","59.917835","1"],["5.359971","59.593094","3"],["5.381013","60.334126","90"],["5.487726","62.436531","3"],["5.507941","59.795212","7"],["5.545053","58.572147","3"],["5.618718","58.999573","1"],["5.621221","59.374187","4"],["5.623686","58.778824","6"],["5.666182","59.118542","44"],["5.745896","58.723969","3"],["5.774172","62.660698","2"],["5.784238","60.386955","2"],["5.899172","58.373005","4"],["5.937029","59.665306","1"],["5.995595","59.936134","2"],["6.033874","62.364449","4"],["6.049094","59.009083","2"],["6.089259","58.638355","3"],["6.091209","62.081760","1"],["6.101040","53.106812","6"],["6.108501","60.395718","1"],["6.145587","61.703308","1"],["6.197891","62.784218","24"],["6.292152","62.261089","2"],["6.312727","58.835342","2"],["6.318905","58.996174","11"],["6.352768","59.651001","1"],["6.472209","58.528709","1"],["6.549636","60.701099","6"],["6.686110","60.135174","3"],["6.687324","62.974514","2"],["6.717930","61.902618","1"],["6.788066","61.297432","2"],["6.816438","59.571545","1"],["6.877266","62.138996","2"],["6.916443","60.567886","1"],["6.978507","58.525082","2"],["7.015129","62.603916","1"],["7.144339","58.149044","1"],["7.182045","60.855564","1"],["7.405971","61.593742","1"],["7.436391","57.880684","3"],["7.475951","63.053268","2"],["7.489135","63.218475","6"],["7.764776","59.889973","1"],["7.809655","58.339207","8"],["7.857247","61.333427","1"],["7.859083","62.748165","7"],["78.486671","17.385044","6"],["8.019276","58.028095","35"],["8.241149","62.915730","1"],["8.302487","56.544346","1"],["8.375780","58.249401","2"],["8.459405","55.476467","1"],["8.546165","58.590969","1"],["8.561080","60.630310","1"],["8.594430","58.342098","4"],["8.764686","58.459755","15"],["8.774252","62.598881","3"],["8.807243","63.546726","2"],["8.880215","62.876709","4"],["8.888742","63.970516","3"],["8.930381","58.622944","4"],["9.080902","60.505749","1"],["9.209406","59.656658","4"],["9.211516","63.056339","1"],["9.253851","60.962791","4"],["9.321380","58.698257","1"],["9.412651","61.841522","1"],["9.425592","60.130840","1"],["9.578526","59.003071","3"],["9.582653","63.782791","3"],["9.602032","59.535206","5"],["9.605492","58.792782","4"],["9.612769","59.188160","12"],["9.655912","59.139706","19"],["9.658694","61.593613","1"],["9.665666","62.587173","1"],["9.679235","60.205914","2"],["9.851449","61.296696","1"],["9.851893","59.796043","3"],["9.921747","57.048820","2"],["9.980854","60.300396","3"]]`
	if ok {
		resp, err := json.Marshal(value.([][]string))
		if err == nil {
			jsonResponse = string(resp)
		}
	}
	fmt.Fprint(w, jsonResponse)
}

func RtDeviceHandler(w http.ResponseWriter, r *http.Request) {
	setCacheHeadersForHTTPHandlers(w)

	var jsonResponse string
	jsonResponse = `[["DESKTOP","0"],["MOBILE","0"],["TABLET","0"]]`
	value, ok := db.Load(RTDevices)
	if ok {
		resp, err := json.Marshal(value.([][]string))
		if err == nil {
			jsonResponse = string(resp)
		}
	}
	fmt.Fprint(w, jsonResponse)

}

// setCacheHeadersForHTTPHandlers sents a set of cache-headers for the handlers
// delivering data.
//
// # Use of max-age
//
// The use of max-age on mutable data is unfortunate in instances where data
// should be in sync, in this case we don´t care. In practice a request
// will most often be sendt without caching from most clients.
func setCacheHeadersForHTTPHandlers(w http.ResponseWriter) {
	value, ok := db.Load(LastModified)
	if !ok {
		value = time.Now()
	}
	modified := value.(time.Time)
	w.Header().Set("Cache-Control", "max-age:30, public")
	w.Header().Set("Last-Modified", modified.Format(http.TimeFormat))
	w.Header().Set("Expires", modified.Add(updateFrequency).Format(http.TimeFormat))
}

// RefreshDataHandler is called every minute and refreshes
// the datasource.
func RefreshDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		done := make(chan bool, 1)

		go func() {
			err := RefreshData(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Update failed")
			}
			done <- true
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			fmt.Println("Early exit, timeout 5 seconds")
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Need POST")
	}
}

// RefreshData is called every 1 minute from cloud scheduler.
//
// # Current API protection
//
// The API is currently only protected by its surrounding infrastructure and it
// doesn´t check that its only called by the Google Cloud Scheduler function.
// The concequences of too many calls will a temporary downtime due to too many
// calls to the GA API which is limited to 150 000 calls in a 24 hour period
// (Since we rouond robin 3 views).
//
// We do have some infrastructure protection from cloudflare if we are hit by an
// extreme amount of traffic from a single source, but still there is a real
// risk that resources might be exchausted so that it stops working until the
// next day.
//
// # Higher frequency
//
// The maximum resolution of the cloud scheduler is 1 minute and also incurres
// http overhead when used.  Historicly I have used refresh-rates of 5 seconds,
// however it sometimes gave a busy view and a slower refresh rate is probably
// better. An illusion of change is done by tweening in the D3 library over a
// few seconds.
//
//	0   5                      30
//	|---|-----------------------|
//	   Tween
func RefreshData(ctx context.Context) (err error) {

	var (
		rtCall, geo, devices *analytics.RealtimeData
	)
	var profiles = [...]string{"ga:78449289", "ga:95719958"}
	view := profiles[rand.Intn(len(profiles))]
	log.Println("View ", view)
	defer timeTrack(time.Now(), "Update timing:")

	var ga *analytics.Service
	ga, err = SetupGoogleAnalyticsService(ctx)

	if err != nil {
		db.Store(RTUsers, rtCall.Rows)
		return
	}

	rtCall, err = ga.Data.Realtime.Get(view, "rt:activeUsers").Do()
	if err != nil {
		return
	} else {
		db.Store(RTUsers, rtCall.Rows)
	}

	geo, err = ga.Data.Realtime.Get(view, "rt:activeUsers").Dimensions("rt:longitude,rt:latitude").Do()
	if err != nil {
		return
	} else {
		db.Store(RTGeo, geo.Rows)
	}

	devices, err = ga.Data.Realtime.Get(view, "rt:activeUsers").Dimensions("rt:deviceCategory").Do()
	if err != nil {
		return
	} else {
		db.Store(RTDevices, devices.Rows)

	}

	db.Store(LastModified, time.Now().UTC())

	return
}

// SetupGoogleAnalyticsService authenticates with GA and retrieve a reporting
// object.
func SetupGoogleAnalyticsService(ctx context.Context) (gaClient *analytics.Service, err error) {
	secret, err := GetApplicationSecrets(ctx)
	if err != nil {
		return
	}
	//	jwtConf, err := google.JWTConfigFromJSON(
	//		secret,
	//		analytics.AnalyticsReadonlyScope,
	//)
	//httpClient := jwtConf.Client(ctx)
	//gaClient, err = analytics.New(httpClient)
	gaClient, err = analytics.NewService(ctx, option.WithCredentialsJSON(secret))
	return
}

// GetApplicationSecrets retrieves secrets because we don´t want to use the
// default serviceaccount. The secret contains oath json credentials.
func GetApplicationSecrets(ctx context.Context) (secrets []byte, err error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/908565461144/secrets/ga-dashboard-serviceaccount/versions/1",
	}

	secretResponse, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return
	}
	secrets = secretResponse.Payload.Data
	return
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
