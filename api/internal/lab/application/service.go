package application

import (
	"alexdenkk/labs/internal/lab/domain"
	"alexdenkk/labs/pkg/config"
	"alexdenkk/labs/pkg/token/jwt"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type service struct {
	repository  domain.LabRepository
	agent       domain.AgentAdapter
	agentConfig *config.AgentConfig
}

func New(
	repository domain.LabRepository,
	agent domain.AgentAdapter,
	agentConfig *config.AgentConfig,
) domain.LabService {
	return &service{
		repository:  repository,
		agent:       agent,
		agentConfig: agentConfig,
	}
}

func (service *service) GetForSelf(ctx context.Context, claims *jwt.Claims) ([]domain.Lab, error) {
	labs, err := service.repository.GetByUser(ctx, claims.UserID)

	if err != nil {
		return []domain.Lab{}, errors.New("labs not found")
	}

	return labs, nil
}

func (service *service) Get(ctx context.Context, id uuid.UUID, claims *jwt.Claims) (domain.Lab, error) {
	lab, err := service.repository.Get(ctx, id)

	if err != nil || lab.UserID != claims.UserID {
		return domain.Lab{}, errors.New("lab not found")
	}

	return lab, nil
}

func (service *service) Delete(ctx context.Context, id uuid.UUID, claims *jwt.Claims) error {
	lab, err := service.repository.Get(ctx, id)

	if err != nil || lab.UserID != claims.UserID {
		return errors.New("lab not found")
	}

	err = service.repository.Delete(ctx, id)

	if err != nil {
		return errors.New("error deleting lab")
	}

	return nil
}

// ГЕНЕРАЦИЯ ГЕНЕРАЦИЯ ГЕНЕРАЦИЯ
func (service *service) Generate(ctx context.Context, lab domain.Lab, claims *jwt.Claims) error {
	formatted, err := service.agent.Call(
		ctx,
		domain.AgentRequest{
			Message: lab.Text,
		},
		service.agentConfig.FormatterAccessID,
	)

	if err != nil {
		println(err.Error())
		return errors.New("ошибка форматирования текста")
	}

	var volumes []map[string]string

	if err = json.Unmarshal([]byte(formatted.Message[7:len(formatted.Message)-3]), &volumes); err != nil {
		return errors.New("ошибка декодирования текста")
	}

	content, err := os.ReadFile("template.typ")

	if err != nil {
		return errors.New("ошибка при чтении шаблона")
	}

	tmpl, err := template.New("lab").Parse(string(content))

	if err != nil {
		return errors.New("ошибка при парсинге шаблона")
	}

	var sb strings.Builder

	err = tmpl.Execute(&sb, lab)

	if err != nil {
		println(err.Error())
		return errors.New("ошибка генерации шаблона")
	}

	firstPage := sb.String()
	imagesForAgent := []map[string]string{}

	for _, img := range lab.Images {
		data, err := base64.StdEncoding.DecodeString(img.Base64)
		if err != nil {
			return err
		}

		fileName := uuid.New().String() + ".png"
		filePath := filepath.Join("files/", fileName)

		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}

		imagesForAgent = append(imagesForAgent, map[string]string{
			"description": img.Description,
			"filename":    fileName,
		})
	}

	encodedLab, _ := json.Marshal(map[string]interface{}{
		"lab":    volumes,
		"images": imagesForAgent,
	})

	resp, err := service.agent.Call(
		ctx,
		domain.AgentRequest{
			Message: string(encodedLab),
		},
		service.agentConfig.GeneratorAccessID,
	)

	if err != nil {
		return err
	}

	preText := strings.Split(resp.Message, "```typst")[1]

	text := firstPage + "\n" + preText[:len(preText)-3]

	lab.ID = uuid.New()
	println(lab.ID.String())

	os.WriteFile("files/"+lab.ID.String()+".typ", []byte(text), 0644)

	err = service.runCommand("./typst", "compile", "files/"+lab.ID.String()+".typ")

	output := ""
	parentID := resp.ID
	generated := false

	if err != nil {
		output = err.Error()
		for i := 0; i <= 2; i++ {
			resp, err := service.agent.Call(
				ctx,
				domain.AgentRequest{
					Message:         output,
					ParentMessageID: parentID,
				},
				service.agentConfig.GeneratorAccessID,
			)

			if err != nil {
				return err
			}

			preText := strings.Split(resp.Message, "```typst")[1]

			text := firstPage + "\n" + preText[:len(preText)-3]

			lab.ID = uuid.New()
			println(lab.ID.String())

			os.WriteFile("files/"+lab.ID.String()+".typ", []byte(text), 0644)

			err = service.runCommand("./typst", "compile", "files/"+lab.ID.String()+".typ")

			if err != nil && !service.fileExists(lab.ID.String()+".pdf") {
				output = err.Error()
				parentID = resp.ID
				continue
			}

			generated = true
		}
	}

	if !generated {
		return errors.New("ошибка при генерации pdf")
	}

	lab.UserID = claims.UserID
	lab.Filename = lab.ID.String()

	err = lab.Validate()

	if err != nil {
		return err
	}

	lab.ID = uuid.New()
	err = service.repository.Create(ctx, lab)

	if err != nil {
		return errors.New("error creating lab")
	}

	return nil
}

func (service *service) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (service *service) runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, stderr.String())
		}
		return err
	}

	return nil
}
