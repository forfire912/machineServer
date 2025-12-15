from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="machineserver",
    version="0.1.0",
    author="MachineServer Team",
    description="统一仿真微服务平台 - A unified simulation microservice platform",
    long_description=long_description,
    long_description_content_type="text/markdown",
    packages=find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.8",
    install_requires=[
        "Flask>=3.0.0",
        "Flask-CORS>=4.0.0",
        "Flask-RESTful>=0.3.10",
        "requests>=2.31.0",
        "pyyaml>=6.0.1",
    ],
)
