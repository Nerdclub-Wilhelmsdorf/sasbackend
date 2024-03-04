import sys
import bcrypt
import secrets
import string
import json
if len(sys.argv) > 2:
    pin = sys.argv[1]
    name = sys.argv[2]
    if len(sys.argv) < 4:
        guest = "false"
    else:
        guest = sys.argv[3]
    hashed = bcrypt.hashpw(pin.encode('utf-8'), bcrypt.gensalt())
    #print(hashed.decode('utf-8'))
    #write pin and name to file named after random id with 16 characters
    id = ''.join(secrets.choice(string.ascii_uppercase + string.ascii_lowercase  + string.digits) for _ in range(16))
    if guest == "guest": 
        with open("documents/"+ id+ ".json", 'w') as file:
            json.dump({"name": name, "pin": hashed.decode('utf-8'), "balance" : "0", "guest" : "true"}, file)
    else:
        with open("documents/"+ id+ ".json", 'w') as file:
            json.dump({"name": name, "pin": hashed.decode('utf-8'), "balance" : "0"}, file)
    print("Account created with id: " + id)


else:
    print("No command line arguments provided.")