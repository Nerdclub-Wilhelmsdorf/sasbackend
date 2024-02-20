import sys
import bcrypt
import random
import string
import json
if len(sys.argv) > 2:
    pin = sys.argv[1]
    name = sys.argv[2]
    hashed = bcrypt.hashpw(pin.encode('utf-8'), bcrypt.gensalt())
    #print(hashed.decode('utf-8'))
    #write pin and name to file named after random id of length 16
    id = ''.join(random.choices(string.ascii_letters + string.digits, k=16))
    with open("documents/"+ id+ ".json", 'w') as file:
        json.dump({"name": name, "pin": hashed.decode('utf-8'), "balance" : "0"}, file)
    print("Account created with id: " + id)


else:
    print("No command line arguments provided.")