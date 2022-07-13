package servidor

import (
	"crud/banco"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

type usuario struct {
	ID uint32 `json:"id"`
	Nome string `json:"nome"`
	Email string `json:"email"`
}

func CriarUsuario(w http.ResponseWriter, r *http.Request) {
	corpoRequisicao, erro := ioutil.ReadAll(r.Body) //Leitura da entrada e saída de dados (corpo da requisição)
	if erro != nil {
		w.Write([]byte("Falha ao ler o corpo da requisição."))
		return
	}

	var usuario usuario //Criando meu usuario em branco

	if erro = json.Unmarshal(corpoRequisicao, &usuario); erro != nil { //Preencher o usuario
		w.Write([]byte("Erro ao converter o usuário para struct."))
		return
	}

	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar com o banco de dados."))
		return
	}
	defer db.Close()

	//PREPARE STATEMENT
	statement, erro := db.Prepare("insert into usuarios (nome, email) values (?,?)") //Não passo valor para não correr risco de ataque
	if erro != nil {
		w.Write([]byte("Erro ao criar o statement"))
		return
	}
	defer statement.Close()

	insercao, erro := statement.Exec(usuario.Nome, usuario.Email) //Substituir os valores para substituir os "?"
	if erro != nil {
		w.Write([]byte("Erro ao executar o statement"))
		return
	}

	//A PARTIR DAQUI É SINAL QUE O USUÁRIO FOI INSERIDO

	idInserido, erro := insercao.LastInsertId()
	if erro != nil {
		w.Write([]byte("Erro ao obter o id inserido."))
		return
	}

	//	STATUS CODE

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Usuario inserido com sucesso! ID: %d", idInserido)))
}

//Buscar todos os usuários (só como consulta)
func BuscarUsuarios(w http.ResponseWriter, r *http.Request) {
	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar com o banco de dados."))
		return
	}
	defer db.Close()

	linhas, erro := db.Query("select * from usuarios")
	if erro != nil {
		w.Write([]byte("Erro ao buscar os usuários."))
		return
	}
	defer linhas.Close()

	var usuarios []usuario //slice de usuarios - podendo retornar mais de uma linha
	for linhas.Next() { //A cada linha retornada ele vai executar uma ação
		var usuario usuario //ler os dados da linha, jogar na struct de usuário e depois jogar esse usuário no slice de usuário

		if erro := linhas.Scan(&usuario.ID, &usuario.Nome, &usuario.Email); erro != nil { //manter a ordem do banco de dados
			w.Write([]byte("Erro ao escanear o usuário."))
			return
		}

		usuarios = append(usuarios, usuario)
	}

	w.WriteHeader(http.StatusOK)
	//codificar os dados pra Json
	if erro := json.NewEncoder(w).Encode(usuarios); erro != nil {
		w.Write([]byte("Erro ao converter os usuários para JSON."))
		return
	}
}

//Buscar um usuário específico no banco
func BuscarUsuario(w http.ResponseWriter, r *http.Request) {

	//Retornando os parametros r baseado no nome que passei no router ( {id} )
	parametros := mux.Vars(r)

	ID, erro := strconv.ParseUint(parametros["id"], 10, 32) //ParseUint recebe 3 parametros: parametro, base do numero (decimal), tamanho dos bits (32)
	if erro != nil {
		w.Write([]byte("Erro ao converter o parâmetro para inteiro."))
		return
	}

	//Vou abrir a conexão com o banco só agora porque já tenho o ID

	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar com o banco de dados."))
		return
	}

	linha, erro := db.Query("select * from usuarios where id = ?", ID)
	if erro != nil {
		w.Write([]byte("Erro ao buscar o usuário."))
		return
	}

	//Criar um usuario só
	var usuario usuario
	if linha.Next() {
		if erro := linha.Scan(&usuario.ID, &usuario.Nome, &usuario.Email); erro != nil {
			w.Write([]byte("Erro ao escanear usuário."))
			return
		}
	}

	if erro := json.NewEncoder(w).Encode(usuario); erro != nil {
		w.Write([]byte("Erro ao converter o usuário para JSON."))
		return
	}
}

//Atualizar um cadastro de usuário no banco de dados
func AtualizarUsuario(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r) //Passando o request

	ID, erro := strconv.ParseUint(parametros["ID"], 10, 32) //converter o parâmetro para inteiro
	if erro != nil {
		w.Write([]byte("Erro ao converter o parâmetro para inteiro."))
		return
	}

	//ler o corpo da requisição e depois abrir o banco (primeiro os requisitos)
	corpoRequisicao, erro := ioutil.ReadAll(r.Body)
	if erro != nil {
		w.Write([]byte("Erro ao ler o corpo da requisição."))
		return
	}

	//Passando o usuario de JSON para struct
	var usuario usuario
	if erro := json.Unmarshal(corpoRequisicao, &usuario); erro != nil {
		w.Write([]byte("Erro ao converter o usuário para a struct"))
		return
	}

	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar com o Banco."))
		return
	}
	defer db.Close()

	statement, erro := db.Prepare("update usuarios set nome=?, email=? where id=?")
	if erro != nil {
		w.Write([]byte("Erro ao criar o statement."))
		return
	}
	defer statement.Close()

	if _, erro := statement.Exec(usuario.Nome, usuario.Email, ID); erro != nil { //o ID é passado dessa maneira porque estou lendo ele no parametro
		w.Write([]byte("Erro ao atualizar o usuário."))
		return
	}

	//Informação no cabeçalho
	w.WriteHeader(http.StatusNoContent) //204


	//Statement usado para qualquer operação que não seja de consulta
}

//Deletar usuário
func DeletarUsuario(w http.ResponseWriter, r *http.Request) {
	parametros := mux.Vars(r)

	ID, erro := strconv.ParseUint(parametros["ID"], 10, 32)
	if erro != nil {
		w.Write([]byte("Erro ao converter o parâmetro para inteiro."))
		return
	}

	db, erro := banco.Conectar()
	if erro != nil {
		w.Write([]byte("Erro ao conectar com o banco da dados."))
		return
	}
	defer db.Close()

	statement, erro := db.Prepare("delete from usuarios where id = ?")
	if erro != nil{
		w.Write([]byte("Erro ao criar o statement."))
		return
	}
	defer statement.Close()

	if _, erro := statement.Exec(ID); erro != nil {
		w.Write([]byte("Erro ao deletar usuário."))
		return
	}

	w.WriteHeader(http.StatusNoContent)

}