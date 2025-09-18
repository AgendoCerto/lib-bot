# Conversão básica (substitui o arquivo existente)
go run main.go -in=bot_atendimento_padronizado.json -out=reactflow

# Com layout vertical
go run main.go -in=bot_atendimento_padronizado.json -out=reactflow-auto-v

# Com layout horizontal  
go run main.go -in=bot_atendimento_padronizado.json -out=reactflow-auto-h

# Nome customizado
go run main.go -in=input.json -out=reactflow -outfile=meu-flow.json