import sys
import bcrypt

if len(sys.argv) > 1:
    first_arg = sys.argv[1]
    hashed = bcrypt.hashpw(first_arg.encode('utf-8'), bcrypt.gensalt())
    print(hashed.decode('utf-8'))
else:
    print("No command line arguments provided.")