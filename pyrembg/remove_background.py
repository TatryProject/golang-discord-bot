from rembg import remove

input_path = '../resized-emote'
output_path = '../rembg-emote'

with open(input_path, 'rb') as i:
    with open(output_path, 'wb') as o:
        input = i.read()
        output = remove(input)
        o.write(output)