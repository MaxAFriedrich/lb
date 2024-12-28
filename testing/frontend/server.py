from pathlib import Path

import requests
import yaml
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()


def get_backend_map():
    backend_map_path = Path(
        __file__).parent.parent / "backend" / "backend-map.yml"
    with open(backend_map_path) as f:
        return yaml.load(f, Loader=yaml.FullLoader)


backend_map = get_backend_map()

lb_endpoint = backend_map["lb_endpoint"]


def build_layout_html() -> str:
    out = "<section>"
    for box in backend_map["layout"]:
        box_id = box["id"]
        box_name = box["name"]

        out += f"<h2>{box_name}</h2><ul>"

        for service in box["services"]:
            service_id = service["id"]
            service_name = service["name"]
            service_description = service["description"]
            service_endpoint = (
                    "/proxy/{{instance_id}}/" +
                    box_id + "/" +
                    service_id
            )

            out += (f"<li><a href='{service_endpoint}' target='_blank'>"
                    f"{service_name}</a><br>{service_description}</li>")

        out += "</ul>"
    out += "</section>"
    return out


def build_index_html() -> str:
    layout_html = build_layout_html()

    out = ("<!DOCTYPE html><html><head>"
           "<title>Load Balancer Test Frontend</title>"
           "<style>body{background: #222; color: #ddd; font-family: "
           "sans-serif;}"
           "section{margin: 20px; padding: 20px; background: #333; "
           "border-radius: 10px;}"
           "h1{color: #fff;}"
           "h2{color: #fff;}"
           "</style>"
           "</head><body>")
    for instance in backend_map["backends"]:
        out += f"<h1>Instance {instance['id']}</h1>"
        instance_id = instance["id"]
        out += layout_html.replace(
            "{{instance_id}}", str(instance_id)
        )

    out += "</body></html>"
    return out


index_html = build_index_html()


@app.get("/", response_class=HTMLResponse)
def read_root():
    return HTMLResponse(content=index_html)


# proxy endpoints are in format /proxy/instance_id/box_id/service_id
@app.get("/proxy/{instance_id}/{box_id}/{service_id}")
def proxy(instance_id: str, box_id: str, service_id: str,
          response_class=HTMLResponse):
    res = requests.get(f"{lb_endpoint}/{instance_id}/{box_id}/{service_id}")
    # TODO make this into a real http proxy
    return HTMLResponse(content=res.text)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8080)
