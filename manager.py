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

    def build_image(self, platform='linux/amd64,linux/arm64', buildx=True, push=False):
        build_params = f'-t {image} {self.root_dir}'
        if buildx:
            self.shell_run('docker buildx create --use --name multi-arch-builder', exit_on_error=False)
            cmd = f'docker buildx build --platform {platform} {build_params}'
            if push:
                cmd += f' --push'
        else:
            cmd = f'docker build {build_params}'
        self.shell_run(cmd)

    # lvps: label value pair str
    def restart(self, metrics_url, pushgateway_base_url, pushgateway_job, log_level='error', pushgateway_username='', pushgateway_password='', lvps='', pushgateway_crontab='*/15 * * * * *', container=None, container_prefix=None):
        if not container:
            container = self._container(metrics_url=metrics_url, pushgateway_job=pushgateway_job)
        if container_prefix:
            container = f'{container_prefix}-{container}'
        self.stop_container(container, timeout=1)
        self.remove_container(container, force=True)

        pushgateway_url_list = []
        pushgateway_url_list.append(pushgateway_base_url)
        pushgateway_url_list.append(f'job/{pushgateway_job}')
        if len(lvps) > 0:
            pushgateway_url_list.append(lvps)

        pushgateway_url = '/'.join(pushgateway_url_list)

        envs = [
            f'MP_LOG_LEVEL="{log_level}"',
            f'MP_METRICS_URL="{metrics_url}"',
            f'MP_PG_URL="{pushgateway_url}"',
            f'MP_PG_USERNAME="{pushgateway_username}"',
            f'MP_PG_PASSWORD="{pushgateway_password}"',
            f'MP_PG_CRONTAB="{pushgateway_crontab}"',
        ]

        args = dockerutil.base_docker_args(container_name=container, envs=envs, auto_hostname=False)

        cmd_data = {'image': image, 'args': args}
        cmd = template.render_str('docker run -d --restart always {{ args }} {{ image }}', cmd_data)
        self.shell_run(cmd)

    def _container(self, metrics_url, pushgateway_job):
        url_com = urlparse(metrics_url)
        metrics_id = f'{url_com.hostname}_{url_com.port}{url_com.path.replace("/", "_")}__{url_com.query.replace("=", "_").replace(":", "_")}'
        container = f'{self.app_name}-{pushgateway_job}-{metrics_id}'
        return container


if __name__ == '__main__':
    functocli.run_app(App)
