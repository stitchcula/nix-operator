{{.CommentHeader}}[connection]
id={{.Interface.Name}}
type=ethernet
interface-name={{.Interface.Name}}

[ipv4]
{{- if .Interface.IPAddress}}
address1={{.Interface.IPAddress}}
method=manual
{{- if .Interface.Gateway}}
gateway={{.Interface.Gateway}}
{{- end}}
{{- else}}
method=disabled
{{- end}}
{{- if .Interface.Nameservers}}
dns={{join .Interface.Nameservers ";"}}
{{- end}}

[ipv6]
{{- if .Interface.IPv6Address}}
address1={{.Interface.IPv6Address}}
method=manual
{{- if .Interface.IPv6Gateway}}
gateway={{.Interface.IPv6Gateway}}
{{- end}}
{{- else}}
method=disabled
{{- end}}