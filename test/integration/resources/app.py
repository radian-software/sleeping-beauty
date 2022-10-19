import flask

app = flask.Flask(__name__)


@app.route("/about")
def route_about():
    return "About this application"
