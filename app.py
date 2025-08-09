
from flask import Flask, request, jsonify
from transformers import pipeline

app = Flask(__name__)

classifier = pipeline("text-classification", model="j-hartmann/emotion-english-distilroberta-base", top_k=None)

@app.route('/analyze_emotion', methods=['POST'])
def analyze_emotion():
    
    data = request.get_json(force=True)
    text = data.get('text', '')

    if not text:
        return jsonify({"error": "No text provided"}), 400

    
    prediction = classifier(text)

    
    return jsonify(prediction), 200

# Run the app
if __name__ == '__main__':
    # The server will run on http://127.0.0.1:5000/
    app.run(debug=True)