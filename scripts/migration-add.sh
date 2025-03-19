#!/bin/bash

echo -n "Enter migration name: "
read -r NAME

goose -dir ./migrations create "$NAME" sql