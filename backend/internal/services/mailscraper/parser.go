package mailscraper

import (
	"strconv"
	"strings"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
)

// OrderMailBody renders an order as the structured mail body the parser
// reads back — the reference documentation of the [ORDINE] mail format
// (also handy for send-test tooling). Keys accept Italian and English.
func OrderMailBody(o dto.InboundOrderResponse) string {
	portal := "no"
	if o.Portal {
		portal = "si"
	}
	return strings.Join([]string{
		"CLIENTE: " + o.Client,
		"MITTENTE: " + o.SenderEmail,
		"RIF: " + o.Ref,
		"PRODOTTO: " + o.Product,
		"KG: " + strconv.Itoa(o.Kg),
		"CARICO: " + o.LoadDate + " | " + o.LoadPlace,
		"CONSEGNA: " + o.DeliveryDate + " | " + o.DeliveryPlace,
		"NOLO: " + o.Rate,
		"PORTALE: " + portal,
		"NOTE: " + o.Notes,
	}, "\n")
}

// parseOrderBody extracts an inbound-order request from a mail body.
// fromAddr is the actual SMTP sender, used when the body carries no
// MITTENTE/SENDER field. Returns ok=false when the body doesn't look like
// an order (no client, or neither ref nor product).
func parseOrderBody(body, fromAddr string) (dto.InboundOrderRequest, bool) {
	fields := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		k, v, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(k))
		val := strings.TrimSpace(v)
		switch key {
		case "CLIENTE", "CLIENT":
			fields["client"] = val
		case "MITTENTE", "SENDER", "FROM":
			fields["sender"] = val
		case "RIF", "REF", "RIFERIMENTO", "REFERENCE":
			fields["ref"] = val
		case "PRODOTTO", "PRODUCT", "MERCE":
			fields["product"] = val
		case "KG", "QTY", "QUANTITA", "QUANTITÀ", "QUANTITY":
			fields["kg"] = val
		case "CARICO", "LOAD", "LOADING":
			fields["load"] = val
		case "CONSEGNA", "SCARICO", "DELIVERY", "UNLOAD":
			fields["delivery"] = val
		case "NOLO", "RATE", "PRICE", "PREZZO":
			fields["rate"] = val
		case "PORTALE", "PORTAL":
			fields["portal"] = strings.ToLower(val)
		case "NOTE", "NOTES":
			fields["notes"] = val
		}
	}
	if fields["client"] == "" || (fields["ref"] == "" && fields["product"] == "") {
		return dto.InboundOrderRequest{}, false
	}

	sender := fields["sender"]
	if sender == "" {
		sender = fromAddr
	}
	kg := 0
	if n, err := strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(fields["kg"], ".", ""), ",", "")); err == nil {
		kg = n
	}
	loadDate, loadPlace := splitDatePlace(fields["load"])
	delDate, delPlace := splitDatePlace(fields["delivery"])

	return dto.InboundOrderRequest{
		Client:        fields["client"],
		SenderEmail:   sender,
		Ref:           fields["ref"],
		Product:       fields["product"],
		Kg:            kg,
		LoadDate:      loadDate,
		LoadPlace:     loadPlace,
		DeliveryDate:  delDate,
		DeliveryPlace: delPlace,
		Rate:          fields["rate"],
		Notes:         fields["notes"],
		Portal:        fields["portal"] == "si" || fields["portal"] == "yes" || fields["portal"] == "true",
		Source:        models.InboundOrderSourceMail,
	}, true
}

func splitDatePlace(v string) (date, place string) {
	if d, p, found := strings.Cut(v, "|"); found {
		return strings.TrimSpace(d), strings.TrimSpace(p)
	}
	return strings.TrimSpace(v), ""
}
