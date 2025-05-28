{{.CommentHeader}}auto {{.Interface.Name}}
{{- if .Interface.IPAddress}}
iface {{.Interface.Name}} inet static
    address {{.Interface.IPAddress}}
{{- if .Interface.Gateway}}
    gateway {{.Interface.Gateway}}
{{- end}}
{{- if .Interface.Nameservers}}
    dns-nameservers {{join .Interface.Nameservers " "}}
{{- end}}
{{- end}}
{{- if .Interface.IPv6Address}}
iface {{.Interface.Name}} inet6 static
    address {{.Interface.IPv6Address}}
{{- if .Interface.IPv6Gateway}}
    gateway {{.Interface.IPv6Gateway}}
{{- end}}
{{- if and .Interface.Nameservers (not .Interface.IPAddress)}}
    dns-nameservers {{join .Interface.Nameservers " "}}
{{- end}}
{{- end}}
    mtu {{.Interface.MTU}}