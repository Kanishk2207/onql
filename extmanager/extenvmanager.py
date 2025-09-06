import os
import json
import subprocess
import docker

class ExtEnvManager:
    """
    Manager for building and running extension components based on a manifest.
    """

    def __init__(self,
                 image_tag: str = 'multi-lang:latest',
                 extensions_root: str = 'extensions'):
        """
        Initialize with the Docker image tag and root directory for extensions.

        :param image_tag: Docker image to use for containers
        :param extensions_root: Root folder where extensions live
        """
        self.client = docker.from_env()
        self.image_tag = image_tag
        self.extensions_root = extensions_root

    def load_manifest(self, ext_name: str) -> dict:
        """
        Load and parse the manifest.json for the given extension.

        :param ext_name: Folder name under extensions_root
        :return: Parsed manifest dictionary
        """
        manifest_path = os.path.join(
            self.extensions_root, ext_name, 'manifest.json'
        )
        with open(manifest_path, 'r') as f:
            return json.load(f)

    def build_extension(self, ext_name: str):
        """
        Build all components of the extension by installing dependencies or compiling.

        :param ext_name: Folder name under extensions_root
        """
        manifest = self.load_manifest(ext_name)
        for comp in manifest.get('components', []):
            name = comp['name']
            lang = comp['language']
            path = comp['path'].rstrip('/')
            dep_file = comp.get('dependencyFile', '')
            comp_dir = os.path.join(self.extensions_root, ext_name, path)

            print(f"Building component '{name}' ({lang})...")
            if lang == 'python':
                venv_dir = os.path.join(comp_dir, '.venv')
                subprocess.run(['python3', '-m', 'venv', venv_dir], check=True)
                pip = os.path.join(venv_dir, 'bin', 'pip')
                req = os.path.join(comp_dir, dep_file)
                subprocess.run([pip, 'install', '--upgrade', 'pip'], check=True)
                subprocess.run([pip, 'install', '-r', req], check=True)

            elif lang == 'nodejs':
                subprocess.run(['npm', 'ci'], cwd=comp_dir, check=True)

            elif lang == 'go':
                subprocess.run(['go', 'mod', 'tidy'], cwd=comp_dir, check=True)
                build_dir = os.path.join(comp_dir, 'build')
                os.makedirs(build_dir, exist_ok=True)
                entry = comp['entry']
                out_bin = os.path.join(build_dir, name)
                subprocess.run(['go', 'build', '-o', out_bin, entry], cwd=comp_dir, check=True)

            elif lang == 'java':
                subprocess.run(['mvn', 'package'], cwd=comp_dir, check=True)

            else:
                raise ValueError(f"Unsupported language: {lang}")

    def run_extension(self, ext_name: str):
        """
        Launch containers for each component defined in the extension's manifest.

        :param ext_name: Folder name under extensions_root
        """
        manifest = self.load_manifest(ext_name)
        for comp in manifest.get('components', []):
            name = comp['name']
            lang = comp['language']
            path = comp['path'].rstrip('/')
            entry = comp['entry']

            container_name = f"{ext_name}-{name}"
            host_path = os.path.abspath(
                os.path.join(self.extensions_root, ext_name, path)
            )
            container_path = f"/extensions/{ext_name}/{path}"

            if lang == 'python':
                cmd = [f"{container_path}/.venv/bin/python", f"{container_path}/{entry}"]
            elif lang == 'nodejs':
                cmd = ['node', f"{container_path}/{entry}"]
            elif lang == 'go':
                cmd = [f"{container_path}/build/{name}"]
            elif lang == 'java':
                cmd = ['java', '-jar', f"{container_path}/{name}.jar"]
            else:
                raise ValueError(f"Unsupported language: {lang}")

            volumes = {host_path: {'bind': container_path, 'mode': 'ro'}}
            print(f"Starting component '{name}' of extension '{ext_name}'...")
            self.run_container(
                name=container_name,
                command=cmd,
                volumes=volumes,
                working_dir=container_path
            )

    def run_container(self,
                      name: str = None,
                      command=None,
                      detach: bool = True,
                      tty: bool = True,
                      **kwargs):
        """
        Run a container from the specified image with optional parameters.

        :param name: Name for the container
        :param command: Command list or string to execute
        :param detach: Detached mode
        :param tty: Allocate TTY
        :param kwargs: Additional docker-py run() parameters
        :return: The created Container object
        """
        print(f"Running container '{name}' from image '{self.image_tag}'...")
        container = self.client.containers.run(
            image=self.image_tag,
            command=command,
            name=name,
            detach=detach,
            tty=tty,
            **kwargs
        )
        print(f"Container started: {container.id[:12]}")
        return container


if __name__ == '__main__':
    manager = ExtEnvManager()
    # Build and run an example extension
    manager.build_extension('polyglot-ext')
    manager.run_extension('polyglot-ext')
