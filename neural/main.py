from pyforest import nltk
from nltk.corpus import stopwords as nltk_stopwords
from tensorflow.keras.models import Sequential
from keras.models import load_model
from tensorflow.keras.layers import Dense, Embedding, LSTM, Dropout, SpatialDropout1D
from tensorflow.keras import utils
from tensorflow.keras.preprocessing.sequence import pad_sequences
from tensorflow.keras.preprocessing.text import Tokenizer
from tensorflow.keras.callbacks import ModelCheckpoint
import pandas as pd
pd.options.mode.chained_assignment = None
import numpy as np
import re
import pika
import matplotlib.pyplot as plt
import json
import requests
import tensorflow as tf
import os
os.environ['TF_FORCE_GPU_ALLOW_GROWTH'] = 'true'

# physical_devices = tf.config.list_physical_devices('GPU')
# tf.config.experimental.set_memory_growth(physical_devices[0], enable=True)

num_words = 2000
embed_dim = 128
lstm_out = 64
max_len = 50

data = pd.read_csv('labeled_data/labeled.csv', header=None, names=['Comment', 'Label'])

positive = pd.read_csv('labeled_data/positive.csv', sep=';', header=None)
negative = pd.read_csv('labeled_data/negative.csv', sep=';', header=None)
positive_text = pd.DataFrame(positive.iloc[:, 3])
negative_text = pd.DataFrame(negative.iloc[:, 3])
positive_text['label'] = [1] * positive_text.shape[0]
negative_text['label'] = [0] * negative_text.shape[0]
labeled_data = pd.concat([positive_text, negative_text])
labeled_data.index = range(labeled_data.shape[0])
labeled_data.columns = ['text', 'label']
comments = pd.concat([data['Comment'], labeled_data['text']])
comments.index = range(comments.shape[0])
y_train = pd.concat([data['Label'], labeled_data['label']])
y_train.index = range(y_train.shape[0])
stopwords = set(nltk_stopwords.words('russian'))


credentials = pika.PlainCredentials('user', 'bitnami')
parameters = pika.ConnectionParameters('localhost', 5672, '/', credentials)
connection = pika.BlockingConnection(parameters)
channel = connection.channel()
channel.queue_declare(queue='comments', durable=True, auto_delete=True)


def clean_comments(text):
    text = text.lower().replace("ё", "е")
    text = re.sub('@[^\s]+', '', text)
    text = re.sub('[^a-zA-Zа-яА-Я]+', ' ', text)
    text = re.sub(' +', ' ', text)
    return text.strip()


def clean_stopwords(text, stopwords):
    text = [word for word in text.split() if word not in stopwords]
    return " ".join(text)


for i in range(0, len(comments)):
    comments[i] = clean_comments(comments[i])
    comments[i] = clean_stopwords(comments[i], stopwords)




tokenizer = Tokenizer()
tokenizer.fit_on_texts(comments)

sequences = tokenizer.texts_to_sequences(comments)

x_train = pad_sequences(sequences, maxlen=max_len)

model = Sequential()
model.add(Embedding(num_words, embed_dim, input_length=max_len))
model.add(SpatialDropout1D(0.1))
model.add(LSTM(lstm_out, dropout=0.2, recurrent_dropout=0.2))
# model.add(Dense(32, activation='relu'))
# model.add(Dropout(0.1))
model.add(Dense(1, activation='sigmoid'))
model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
print(model.summary())

model_save_path = 'best_model.h5'
checkpoint_callback = ModelCheckpoint(model_save_path, monitor='val_accuracy', save_best_only=True, verbose=1)
history = model.fit(x_train, y_train, epochs=7, batch_size=128, validation_split=0.1, callbacks=[checkpoint_callback])

plt.plot(history.history['accuracy'], label='Доля верных ответов на обучающем наборе')
plt.plot(history.history['val_accuracy'], label='Доля верных ответов на проверочном наборе')
plt.xlabel('Эпоха обучения')
plt.ylabel('Доля верных ответов')
plt.legend()
plt.show()
model.load_weights('best_model.h5')

test_data = pd.read_csv('labeled_data/test_labeled.csv', header=None, names=['Comment', 'Label'])
test_comments = test_data['Comment']
for i in range(0, len(test_comments)):
    test_comments[i] = clean_comments(test_comments[i])
    test_comments[i] = clean_stopwords(test_comments[i], stopwords)


test_sequences = tokenizer.texts_to_sequences(test_comments)
x_test = pad_sequences(test_sequences, maxlen=max_len)
y_test = test_data['Label']
model.evaluate(x_test, y_test, verbose=1)


def callback(ch, method, properties, body):
    print(" [x] Received %r" % body.decode('utf-8'))
    raw = json.loads(body.decode('utf-8'))
    text = raw['text']
    text = clean_comments(text)
    text = clean_stopwords(text, stopwords)
    print(text)
    sequence = tokenizer.texts_to_sequences([text])
    data_com = pad_sequences(sequence, maxlen=max_len)
    result = model.predict(data_com)
    print(result)
    if result[[0]] < 0.5:
        print('Комментарий положительный')
    else:
        print('Комментарий отрицательный')
        requests.post(url='http://localhost:8081/delete', json=raw)


channel.basic_consume(queue='comments', auto_ack=True, on_message_callback=callback)
print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()

