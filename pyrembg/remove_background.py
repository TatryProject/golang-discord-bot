from rembg import remove

def remove_background()
    input_path = '../resized-emote'
    output_path = '../rembg-emote'

    with open(input_path, 'rb') as i:
        with open(output_path, 'wb') as o:
            print("We are in the python prog")
            input = i.read()
            output = remove(input)
            o.write(output)