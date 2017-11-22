package api

type URLProvider struct {
	context        Context
	dataHolder     *DataHolder
	observedFields map[string]bool
}

func newUrlProvider(c Context, h *DataHolder) (*URLProvider) {
	return &URLProvider{context: c, dataHolder: h, observedFields: map[string]bool{}}
}

func (p *URLProvider) Get(name string) (string, bool) {
	p.observedFields[name] = true
	val, ok := p.dataHolder.Get(p.context, name).(string)
	return val, ok
}

/*func (a *SDK) publish(r *http.Request) {
	ctx := NewContext(r)

	appHostname := appengine.DefaultVersionHostname(ctx.Context)

	var sendValues = map[string]interface{}{
		"email":     holder.GetInput("email").(string),
		"signature": "varanox-admin",
	}

	bs, _ := json.Marshal(sendValues)

	client := urlfetch.Client(ctx.Context)
	resp, err := client.Post("https://"+holder.GetInput("domain").(string)+"/api/auth/client", "application/json", bytes.NewReader(bs))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}
}*/

func (a *SDK) delete() {

}
