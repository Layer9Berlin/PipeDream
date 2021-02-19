package graph

func graphTemplate() string {
	return `
<!DOCTYPE html>
<meta charset="UTF-8">
<html lang="en">
<head>
    <title>Network</title>
    <script
            type="text/javascript"
            src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"
    ></script>
    <style type="text/css">
        #run_graph {
            position: fixed;
            left: 0;
            top: 0;
            right: 0;
            bottom: 0;
        }
    </style>
</head>
<body>
<div id="run_graph"></div>
<script type="text/javascript">
    const nodes = new vis.DataSet([
        {{ range .UserRuns }}{id: {{.Id}}, label: {{ .GraphLabel }}, group: {{ .GraphGroup }} },
		{{ end }}]);

    const edges = new vis.DataSet([
		{{ range .Connections }}{from: {{ .Source.Id }}, to: {{ .Target.Id }}{{ if .Label }}, label: {{ .Label }}{{ end }} },
		{{ end }}]);

    const container = document.getElementById("run_graph");
    const data = {
        nodes: nodes,
        edges: edges,
    };
    const options = {
        edges: {
            arrows: 'middle',
            width: 1.5,
            smooth: true,
            color: '#BBBBBB',
            arrowStrikethrough: false
        },
        groups: {
            active: {
                color: {
                    border: 'rgb(26, 66, 117)',
                    background: 'rgb(164, 237, 250)'
                },
                font: {
                    color: 'rgb(26, 66, 117)'
                }
            },
            error: {
                color: {
                    border: 'rgb(144, 28, 62)',
                    background: 'rgb(248, 206, 170)'
                },
                font: {
                    color: 'rgb(144, 28, 62)'
                }
            },
            success: {
                color: {
                    border: 'rgb(33, 83, 80)',
                    background: 'rgb(187, 238, 187)'
                },
                font: {
                    color: 'rgb(33, 83, 80)'
                }
            },
            warning: {
                color: {
                    border: 'rgb(147, 59, 40)',
                    background: 'rgb(250, 232, 145)'
                },
                font: {
                    color: 'rgb(147, 59, 40)'
                }
            }
        },
        interaction: {
            dragNodes: false,
            dragView: false,
            selectable: false,
            zoomView: false
        },
        nodes: {
            borderWidth: 1.5,
            borderWidthSelected: 1.5,
            color: {
                background: '#E5E5E5',
                border: '#343434'
            },
            shape: "box",
            shapeProperties: {
                borderRadius: 6
            },
            font: {
                color: '#343434',
                multi: true
            }
        },
        physics: {
            enabled: true
        }
    };
    const network = new vis.Network(container, data, options);
</script>
</body>
</html>
`
}
