# Source Code partially taken from: https://github.com/pytorch/tutorials/blob/master/intermediate_source/realtime_rpi.rst
# the above code is licensed under the BSD-3-Clause license (permissive license)

import time

import torch
import numpy as np
from torchvision import models, transforms

import cv2
from PIL import Image
from classes import *

import socket
from datetime import datetime
import random
import json

torch.backends.quantized.engine = 'qnnpack'

cap = cv2.VideoCapture(0, cv2.CAP_V4L2)
cap.set(cv2.CAP_PROP_FRAME_WIDTH, 224)
cap.set(cv2.CAP_PROP_FRAME_HEIGHT, 224)
cap.set(cv2.CAP_PROP_FPS, 36)

preprocess = transforms.Compose([
    transforms.ToTensor(),
    transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225]),
])

net = models.quantization.mobilenet_v2(pretrained=True, quantize=True)
# jit model to take it from ~20fps to ~30fps
net = torch.jit.script(net)

torch.set_num_threads(2)


started = time.time()
last_logged = time.time()
frame_count = 0

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)  # UDP

with torch.no_grad():
    while True:
        # read frame
        ret, image = cap.read()
        if not ret:
            raise RuntimeError("failed to read frame")

        cv2.imshow('WildFogs', image)
        if cv2.waitKey(1) == 27: 
            break  # esc to quit
        
        # convert opencv output from BGR to RGB
        image = image[:, :, [2, 1, 0]]
        permuted = image

        # preprocess
        input_tensor = preprocess(image)

        # create a mini-batch as expected by the model
        input_batch = input_tensor.unsqueeze(0)

        # run model
        output = net(input_batch)
        # do something with output ...

        # log model performance
        frame_count += 1
        now = time.time()
        if now - last_logged > 1:
            # print(f"{frame_count / (now-last_logged)} fps")
            last_logged = now
            frame_count = 0
            
            top = list(enumerate(output[0].softmax(dim=0)))
            top.sort(key=lambda x: x[1], reverse=True)
            # for idx, val in top[:1]:
                # print(f"{val.item()*100:.2f}% {classes[idx]}")
            
            deteced_object = classes[top[0][0]]
            temperature = random.randint(2200, 2300) / 100.0
            dt = str(datetime.now())
            message = str(json.dumps(
                {"detection_time": dt, "detected_object": deteced_object, "temperature": temperature}))
            print(message)
            sock.sendto(message.encode(), ('127.0.0.1', 3333))