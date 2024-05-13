from setuptools import setup, find_packages

NAME = "plugin"
VERSION = "1.0.0"

# To install the library, run the following
#
# python setup.py install
#
# prerequisite: setuptools
# http://pypi.python.org/pypi/setuptools

REQUIRES = [
    "Flask>=3.0.3",
    "python_dateutil>=2.6.0"
]

setup(
    name=NAME,
    version=VERSION,
    description="VMClarity Scanner Plugin SDK",
    url="vmclarity.io",
    keywords=["Scanner Plugin API"],
    install_requires=REQUIRES,
    packages=find_packages(),
    include_package_data=True,
    long_description="SDK to simplify development of scanner plugins used in VMClarity"
)
