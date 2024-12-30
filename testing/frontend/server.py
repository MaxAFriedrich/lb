from pathlib import Path

import httpx
import websockets
import yaml
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from starlette import websockets
from starlette.responses import StreamingResponse
from starlette.websockets import WebSocketDisconnect, WebSocket

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
           "a{color: #fff; text-decoration: none;}"
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

@app.get("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.post("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.put("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.delete("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.patch("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.options("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.head("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
@app.trace("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
async def proxy_http(instance_id: str, box_id: str, service_id: str, path: str,
                     request: Request):
    url = f"{lb_endpoint}/{instance_id}/{box_id}/{service_id}/{path}"
    if request.query_params:
        url += f"?{request.query_params}"
    async with httpx.AsyncClient() as client:
        proxy_request = client.build_request(
            method=request.method,
            url=url,
            headers=request.headers.raw,
            content=await request.body()
        )
        proxy_response = await client.send(proxy_request, stream=True)
        return StreamingResponse(
            proxy_response.aiter_raw(),
            status_code=proxy_response.status_code,
            headers=proxy_response.headers
        )


@app.websocket("/proxy/{instance_id}/{box_id}/{service_id}/{path:path}")
async def proxy_websocket(websocket: WebSocket, instance_id: str, box_id: str,
                          service_id: str, path: str):
    await websocket.accept()
    lb_endpoint_hostname = lb_endpoint.split("://")[1]
    url = (f"ws://{lb_endpoint_hostname}/{instance_id}/{box_id}/{service_id}/"
           f"{path}")
    async with websockets.connect(url) as ws:
        try:
            async for message in websocket.iter_text():
                await ws.send(message)
                response = await ws.recv()
                await websocket.send_text(response)
        except WebSocketDisconnect:
            await ws.close()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8080)
