from urllib.parse import urlparse
from echobox.tool import template
from echobox.tool import functocli
from echobox.tool import dockerutil
from echobox.app.devops import DevOpsApp

APP_NAME = 'metrics-pusher'
image = 'kiuber/metrics-pusher'


class App(DevOpsApp):

    def __init__(self):
        DevOpsApp.__init__(self, APP_NAME)

    def build_image(self, buildx=True, push=False):
        build_params = f'-t {image} {self.root_dir}'
        if buildx:
            self.shell_run('docker buildx create --use --name multi-arch-builder', exit_on_error=False)
            cmd = f'docker buildx build --platform linux/amd64,linux/arm64 {build_params}'
            if push:
                cmd += f' --push'
        else:
            cmd = f'docker build {build_params}'
        self.shell_run(cmd)

    def restart(self, metrics_url, pushgateway_base_url, pushgateway_job, pushgateway_username='', pushgateway_password='', instance_on_metrics=None, pushgateway_crontab='*/15 * * * * *'):
        container = self._container(metrics_url=metrics_url, pushgateway_job=pushgateway_job)
        self.stop_container(container, timeout=1)
        self.remove_container(container, force=True)

        pushgateway_url_list = []
        pushgateway_url_list.append(pushgateway_base_url)
        pushgateway_url_list.append(f'job/{pushgateway_job}')
        if instance_on_metrics:
            pushgateway_url_list.append(f'instance/{instance_on_metrics}')

        pushgateway_url = '/'.join(pushgateway_url_list)

        envs = [
            f'METRICS_URL="{metrics_url}"',
            f'PG_URL="{pushgateway_url}"',
            f'PG_USERNAME="{pushgateway_username}"',
            f'PG_PASSWORD="{pushgateway_password}"',
            f'PG_CRONTAB="{pushgateway_crontab}"',
        ]

        args = dockerutil.base_docker_args(container_name=container, envs=envs, auto_hostname=False)

        cmd_data = {'image': image, 'args': args}
        cmd = template.render_str('docker run -d --restart always {{ args }} {{ image }}', cmd_data)
        self.shell_run(cmd)

    def _container(self, metrics_url, pushgateway_job):
        url_com = urlparse(metrics_url)
        metrics_id = f'{url_com.hostname}_{url_com.port}{url_com.path.replace("/", "_")}__{url_com.query.replace("=", "_")}'
        container = f'{self.app_name}-{pushgateway_job}-{metrics_id}'
        return container


if __name__ == '__main__':
    functocli.run_app(App)
