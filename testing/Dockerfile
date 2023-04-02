FROM python:3

#RUN apt-get update -y
#RUN apt-get install -y python3
#RUN apt-get install -y pip

WORKDIR /testing
COPY . ./


RUN python3 -m pip install --trusted-host pypi.python.org  -r ./requirements.txt

CMD ["python3", "tests.py"]