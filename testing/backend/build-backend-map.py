import re
from dataclasses import dataclass
from typing import Any

import yaml


@dataclass
class Service:
    friendly_name: str
    friendly_description: str
    id: str
    proxy_type: str
    ports: list[int]

    def layout_dict(self):
        return {
            "name": self.friendly_name,
            "description": self.friendly_description,
            "id": self.id,
            "proxy": self.proxy_type,
        }


@dataclass
class Box:
    friendly_name: str
    id: str
    services: list[Service]
    hostname: str

    @property
    def no_instances(self):
        return min([len(service.ports) for service in self.services])

    def layout_dict(self):
        return {
            "name": self.friendly_name,
            "id": self.id,
            "services": [service.layout_dict() for service in self.services],
        }

    def backend_services(self, instance_id):
        out = []
        for service in self.services:
            port = service.ports[instance_id]
            out.append({
                "box_id": self.id,
                "service_id": service.id,
                "host": f"{self.hostname}:{port}",
            })
        return out


@dataclass
class BackendMap:
    lb_endpoint: str
    docker_boxes: list[Box]

    @property
    def __no_instances(self):
        return min([box.no_instances for box in self.docker_boxes])

    @property
    def layout(self):
        return [box.layout_dict() for box in self.docker_boxes]

    @property
    def backends(self):
        instances = []
        for instance_id in range(self.__no_instances):
            services = []
            for box in self.docker_boxes:
                services.extend(box.backend_services(instance_id))
            instances.append({
                "services": services,
                "id": instance_id
            })
        return instances

    def to_dict(self):
        return {
            "lb_endpoint": self.lb_endpoint,
            "layout": self.layout,
            "backends": self.backends
        }


def get_docker_ports(ports: str) -> list[dict[str, str]]:
    regex = r" {6}- \"([\d\.]+):(\d{1,5}):\d{1,5}\" #(\w{1,8})"
    matches = re.findall(regex, ports)
    out = []
    for match in matches:
        out.append({
            "host": match[0],
            "port": match[1],
            "proxy_type": match[2]
        })
    return out


def get_docker_box_instances(compose_file: str) -> list[dict[str, Any]]:
    regex = (r"host(\d)_instance(\d):(\n.*)*?(\n {4}ports:)((\n {6}- \"["
             r"\d:\.]+:\d{1,5}\" #\w{1,8})*)")
    matches = re.findall(regex, compose_file)
    out = []
    for match in matches:
        host_id, instance_id, ports = match[0], match[1], match[4]

        out.append({
            "box_id": host_id,
            "instance_id": instance_id,
            "ports": get_docker_ports(ports)
        })

    return out


def docker_compose_to_backend_map(lb_endpoint: str):
    with open("docker-compose.yml") as f:
        compose_file = f.read()
    instances = get_docker_box_instances(compose_file)
    docker_boxes = []
    for instance in instances:
        box_id = instance["box_id"]
        if not any(box.id == f"box{box_id}" for box in docker_boxes):
            docker_boxes.append(build_box(instance))
        else:
            for box in docker_boxes:
                if box.id == f"box{box_id}":
                    new_services = build_box(instance).services
                    if len(new_services) != len(box.services):
                        raise ValueError(
                            "All boxes must have the same number of services")
                    for i, service in enumerate(new_services):
                        box.services[i].ports.append(service.ports[0])
    return BackendMap(lb_endpoint, docker_boxes)


def build_box(instance) -> Box:
    box_id = instance["box_id"]
    ports = instance["ports"]
    services = []
    hostnames = set()
    for port in ports:
        hostnames.add(port["host"])
        services.append(Service(
            friendly_name=f"Service {port['proxy_type']} on {port['port']}",
            friendly_description="This is a service description",
            id=port["proxy_type"],
            proxy_type=port["proxy_type"],
            ports=[int(port["port"])]
        ))
    if len(hostnames) != 1:
        raise ValueError("All services in a box must have the same hostname")
    return Box(
        friendly_name=f"Host {box_id}",
        id=f"box{box_id}",
        services=services,
        hostname=hostnames.pop()
    )


def main():
    lb_endpoint = "http://localhost:8000"
    backend_map = docker_compose_to_backend_map(lb_endpoint)
    with open("backend-map.yml", "w") as f:
        yaml.dump(backend_map.to_dict(), f)


if __name__ == "__main__":
    main()
